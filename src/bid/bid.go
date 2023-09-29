package bid

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/bet"
	"sammy.link/database"
	"sammy.link/util"
)

type Bid struct {
	Kind             string    `json:"kind"`
	AwayTeam         string    `json:"awayTeam"`
	HomeTeam         string    `json:"homeTeam"`
	ChosenCompetitor string    `json:"chosenCompetitor"`
	Spread           string    `json:"spread"`
	Amount           int64     `json:"amount"`
	Date             time.Time `json:"date"`
	CreateDate       time.Time `json:"createDate"`
	User             string    `json:"user"`
	Week             int       `json:"week"`
	HomeAbbreviation string    `json:"homeAbbreviation"`
	AwayAbbreviation string    `json:"awayAbbreviation"`
	Div              string    `json:"div"`
}

type DyanmoBidItem struct {
	Id               string `dynamodbav:"id"`
	SortKey          string `dynamodbav:"sortKey"`
	Gsi1_id          string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey     string `dynamodbav:"gsi1_sortKey"`
	Spread           string `dynamodbav:"spread"`
	Ttl              int64  `dynamodbav:"ttl"`
	Week             int    `dynamodbav:"we"`
	HomeAbbreviation string `dynamodbav:"h_ab"`
	AwayAbbreviation string `dynamodbav:"a_ab"`
}

type BidAndBet struct {
	MyBids   []Bid
	MyBets   []bet.Bet
	IsDelete bool
	IsBid    bool
}

type Service interface {
	GetBidsByEvent(ctx context.Context, event string, div string) []Bid
	GetBidsByUser(ctx context.Context, user string) []Bid
	Update(ctx context.Context, updateBid Bid)
	WriteBids(ctx context.Context, items []Bid)
	WriteBidsAndBets(ctx context.Context, bidsAndBets []BidAndBet, waitGroup *sync.WaitGroup)
	Lock(ctx context.Context, key string)
	Delete(ctx context.Context, bid Bid)
	ReleaseLock(ctx context.Context, key string)
}
type BidService struct {
	databaseService database.Service[DyanmoBidItem, Bid]
}

func NewService(databaseService database.Service[DyanmoBidItem, Bid]) Service {
	return &BidService{
		databaseService: databaseService,
	}
}

func GetBidDynamoId(div string, kind string, date time.Time, awayTeam string, homeTeam string) string {
	return fmt.Sprintf("B|%s|%s|%s|%s|%s", div, kind, date.Format(time.RFC3339), awayTeam, homeTeam)
}

func GetBidDynamoSortKey(chosenTeam string, user string, createDate time.Time) string {
	return fmt.Sprintf("%s|%s|%d", chosenTeam, user, createDate.Unix())
}

func (bid Bid) GetDynamoItem() database.DynamoItem {
	return DyanmoBidItem{
		Id:               GetBidDynamoId(bid.Div, bid.Kind, bid.Date, bid.AwayTeam, bid.HomeTeam),
		SortKey:          GetBidDynamoSortKey(bid.ChosenCompetitor, bid.User, bid.CreateDate),
		Gsi1_id:          fmt.Sprintf("BID|%s", bid.User),
		Gsi1_sortKey:     strconv.FormatInt(bid.Amount, 10),
		Spread:           bid.Spread,
		Week:             bid.Week,
		AwayAbbreviation: bid.AwayAbbreviation,
		HomeAbbreviation: bid.HomeAbbreviation,
		Ttl:              bid.Date.AddDate(0, 0, 1).Unix(),
	}
}

func (bid DyanmoBidItem) GetItem() database.Item {
	idExpression := regexp.MustCompile(`B\|(?P<Div>[^|]+)\|(?P<Kind>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)`)
	sortKeyExpression := regexp.MustCompile(`(?P<ChosenCompetitor>[^|]+)\|(?P<User>[^|]+)\|(?P<CreateDate>[^|]+)`)

	idMatch := idExpression.FindStringSubmatch(bid.Id)
	sortKeyMatch := sortKeyExpression.FindStringSubmatch(bid.SortKey)

	paramsMap := make(map[string]string)
	for i, name := range idExpression.SubexpNames() {
		if i > 0 && i <= len(idMatch) {
			paramsMap[name] = idMatch[i]
		}
	}

	for i, name := range sortKeyExpression.SubexpNames() {
		if i > 0 && i <= len(sortKeyMatch) {
			paramsMap[name] = sortKeyMatch[i]
		}
	}

	date, _ := time.Parse(time.RFC3339, paramsMap["Date"])
	amount, _ := strconv.ParseInt(bid.Gsi1_sortKey, 10, 64)

	createDateUnix, _ := strconv.ParseInt(paramsMap["CreateDate"], 10, 64)
	createDate := time.Unix(createDateUnix, 0)

	return Bid{
		Kind:             paramsMap["Kind"],
		AwayTeam:         paramsMap["AwayTeam"],
		HomeTeam:         paramsMap["HomeTeam"],
		ChosenCompetitor: paramsMap["ChosenCompetitor"],
		Spread:           bid.Spread,
		Amount:           amount,
		Date:             date,
		CreateDate:       createDate,
		Week:             bid.Week,
		HomeAbbreviation: bid.HomeAbbreviation,
		AwayAbbreviation: bid.AwayAbbreviation,
		User:             paramsMap["User"],
		Div:              paramsMap["Div"],
	}
}

func (s *BidService) Lock(ctx context.Context, key string) {
	s.databaseService.Lock(ctx, key)
}

func (s *BidService) ReleaseLock(ctx context.Context, key string) {
	s.databaseService.ReleaseLock(ctx, key)
}

func (s *BidService) WriteBidsAndBets(ctx context.Context, bidsAndBets []BidAndBet, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	putItems := make([]types.WriteRequest, 0, 25)
	for _, item := range bidsAndBets {
		if item.IsBid {
			if item.IsDelete {
				for _, deleteBid := range item.MyBids {
					putItems = append(putItems, types.WriteRequest{DeleteRequest: &types.DeleteRequest{
						Key: map[string]types.AttributeValue{
							"id": &types.AttributeValueMemberS{
								Value: GetBidDynamoId(deleteBid.Div, deleteBid.Kind, deleteBid.Date, deleteBid.AwayTeam, deleteBid.HomeTeam),
							},
							"sortKey": &types.AttributeValueMemberS{
								Value: GetBidDynamoSortKey(deleteBid.ChosenCompetitor, deleteBid.User, deleteBid.CreateDate),
							},
						},
					}})
				}
			} else {
				attributeValueMap, _ := attributevalue.MarshalMap(item.MyBids[0].GetDynamoItem())
				putItems = append(putItems, types.WriteRequest{PutRequest: &types.PutRequest{
					Item: attributeValueMap,
				}})
			}
		} else {
			for _, newBet := range item.MyBets {
				attributeValueMap, _ := attributevalue.MarshalMap(newBet.GetDynamoItem())

				putItems = append(putItems, types.WriteRequest{PutRequest: &types.PutRequest{
					Item: attributeValueMap,
				}})
			}

		}
	}
	// for _, putItem := range putItems {
	// }
	for i := 0; i < len(putItems); i += 25 {
		waitGroup.Add(1)
		go func(myPutItems []types.WriteRequest) {
			defer waitGroup.Done()
			writeRequests := map[string][]types.WriteRequest{}
			writeRequests[os.Getenv("TABLE_NAME")] = myPutItems

			s.databaseService.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
				RequestItems: writeRequests,
			},
			)
		}(putItems[i:util.Min(i+25, len(putItems))])
	}

}

func (s *BidService) Update(ctx context.Context, updateBid Bid) {
	s.databaseService.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: GetBidDynamoId(updateBid.Div, updateBid.Kind, updateBid.Date, updateBid.AwayTeam, updateBid.HomeTeam)},
			"sortKey": &types.AttributeValueMemberS{Value: GetBidDynamoSortKey(updateBid.ChosenCompetitor, updateBid.User, updateBid.CreateDate)},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberS{Value: strconv.FormatInt(updateBid.Amount, 10)},
		},
		TableName:           aws.String(os.Getenv("TABLE_NAME")),
		UpdateExpression:    aws.String("SET gsi1_sortKey = :amount"),
		ConditionExpression: aws.String("attribute_exists(sortKey)"),
	})
}

func (s *BidService) Delete(ctx context.Context, updateBid Bid) {
	var dynamoBid = updateBid.GetDynamoItem().(DyanmoBidItem)
	fmt.Printf("dynamo %+v\n", dynamoBid)
	s.databaseService.Delete(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: dynamoBid.Id},
			"sortKey": &types.AttributeValueMemberS{Value: dynamoBid.SortKey},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
}

func (s *BidService) GetBidsByEvent(ctx context.Context, event string, div string) []Bid {
	eventExpression := regexp.MustCompile(`(?P<Kind>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)\|?(?P<ChosenCompetitor>[^|]+)?`)

	eventMatch := eventExpression.FindStringSubmatch(event)

	paramsMap := make(map[string]string)
	for i, name := range eventExpression.SubexpNames() {
		if i > 0 && i <= len(eventMatch) && eventMatch[i] != "" {
			paramsMap[name] = eventMatch[i]
		}
	}

	date, _ := time.Parse(time.RFC3339, paramsMap["Date"])

	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: GetBidDynamoId(div, paramsMap["Kind"], date, paramsMap["AwayTeam"], paramsMap["HomeTeam"])},
		},
	})
}

func (s *BidService) GetBidsByUser(ctx context.Context, user string) []Bid {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		IndexName:              aws.String("gsi1"),
		KeyConditionExpression: aws.String("gsi1_id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BID|%s", user)},
		},
	})
}

func (s *BidService) WriteBids(ctx context.Context, items []Bid) {
	s.databaseService.Write(ctx, items)
}
