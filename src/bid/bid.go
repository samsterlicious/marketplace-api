package bid

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
}

type DyanmoBidItem struct {
	Id           string `dynamodbav:"id"`
	SortKey      string `dynamodbav:"sortKey"`
	Gsi1_id      string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey string `dynamodbav:"gsi1_sortKey"`
	Spread       string `dynamodbav:"spread"`
	Ttl          int64  `dynamodbav:"ttl"`
}

func GetBidDynamoId(kind string, date time.Time, awayTeam string, homeTeam string) string {
	return fmt.Sprintf("BID|%s|%s|%s|%s", kind, date.Format(time.RFC3339), awayTeam, homeTeam)
}

func GetBidDynamoSortKey(chosenTeam string, user string, createDate time.Time) string {
	return fmt.Sprintf("%s|%s|%d", chosenTeam, user, createDate.Unix())
}

func (bid Bid) getDynamoItem() DyanmoBidItem {
	return DyanmoBidItem{
		Id:           GetBidDynamoId(bid.Kind, bid.Date, bid.AwayTeam, bid.HomeTeam),
		SortKey:      GetBidDynamoSortKey(bid.ChosenCompetitor, bid.User, bid.CreateDate),
		Gsi1_id:      fmt.Sprintf("BID|%s", bid.User),
		Gsi1_sortKey: strconv.FormatInt(bid.Amount, 10),
		Spread:       bid.Spread,
		Ttl:          bid.Date.AddDate(0, 0, 1).Unix(),
	}
}

func (bid DyanmoBidItem) GetItem() Bid {
	idExpression := regexp.MustCompile(`BID\|(?P<Kind>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)`)
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
		User:             paramsMap["User"],
	}
}

func GetBidsByEvent(ctx context.Context, event string, client *dynamodb.Client) []DyanmoBidItem {
	eventExpression := regexp.MustCompile(`(?P<Kind>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)\|?(?P<ChosenCompetitor>[^|]+)?`)

	eventMatch := eventExpression.FindStringSubmatch(event)

	paramsMap := make(map[string]string)
	for i, name := range eventExpression.SubexpNames() {
		if i > 0 && i <= len(eventMatch) && eventMatch[i] != "" {
			paramsMap[name] = eventMatch[i]
		}
	}

	date, _ := time.Parse(time.RFC3339, paramsMap["Date"])

	return util.Query[DyanmoBidItem](ctx, client, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: GetBidDynamoId(paramsMap["Kind"], date, paramsMap["AwayTeam"], paramsMap["HomeTeam"])},
		},
	})
}

func GetBidsByUser(ctx context.Context, user string, client *dynamodb.Client) []DyanmoBidItem {
	return util.Query[DyanmoBidItem](ctx, client, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		IndexName:              aws.String("gsi1"),
		KeyConditionExpression: aws.String("gsi1_id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BID|%s", user)},
		},
	})
}

func WriteBids(ctx context.Context, items []Bid, client *dynamodb.Client) {

	requestItems := map[string][]types.WriteRequest{}

	putItems := make([]types.WriteRequest, len(items))

	for i, item := range items {
		attributeValueMap, _ := attributevalue.MarshalMap(item.getDynamoItem())

		putItems[i] = types.WriteRequest{PutRequest: &types.PutRequest{
			Item: attributeValueMap,
		}}
	}

	requestItems[os.Getenv("TABLE_NAME")] = putItems

	_, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	},
	)

	if err != nil {
		fmt.Println(err.Error())
	}
}
