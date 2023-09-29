package marketplace

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/bid"
	"sammy.link/database"
)

type MarketplaceItem struct {
	AwayTeam         string    `json:"awayTeam"`
	HomeTeam         string    `json:"homeTeam"`
	AwayAbbreviation string    `json:"awayAbbreviation"`
	HomeAbbreviation string    `json:"homeAbbreviation"`
	AwayRecord       string    `json:"awayRecord"`
	HomeRecord       string    `json:"homeRecord"`
	Id               string    `json:"id"`
	Date             time.Time `json:"date"`
	Kind             string    `json:"kind"`
	Spread           string    `json:"spread"`
	HomeAmount       int64     `json:"homeAmount"`
	AwayAmount       int64     `json:"awayAmount"`
	Week             int       `json:"week"`
}

type MarketplaceDynamoDbItem struct {
	Id               string `dynamodbav:"id"`
	SortKey          string `dynamodbav:"sortKey"`
	Spread           string `dynamodbav:"spread"`
	Ttl              int64  `dynamodbav:"ttl"`
	AwayAbbreviation string `dynamodbav:"awayAb"`
	HomeAbbreviation string `dynamodbav:"homeAb"`
	AwayRecord       string `dynamodbav:"aR"`
	HomeRecord       string `dynamodbav:"hR"`
	EventId          string `dynamodbav:"eId"`
	HomeAmount       int64  `dynamodbav:"homeAmount"`
	AwayAmount       int64  `dynamodbav:"awayAmount"`
	Week             int    `dynamodbav:"week"`
}

type Service interface {
	GetItems(ctx context.Context) []MarketplaceItem
	ModifyAmount(ctx context.Context, bid bid.Bid)
	Write(ctx context.Context, items []MarketplaceItem)
}
type MarketplaceService struct {
	databaseService database.Service[MarketplaceDynamoDbItem, MarketplaceItem]
}

func NewService(databaseService database.Service[MarketplaceDynamoDbItem, MarketplaceItem]) Service {
	return &MarketplaceService{
		databaseService: databaseService,
	}
}

func (item MarketplaceDynamoDbItem) String() string {
	return fmt.Sprintf("id:%s\tsortKey:%s", item.Id, item.SortKey)
}

func (item MarketplaceItem) GetDynamoItem() database.DynamoItem {
	return MarketplaceDynamoDbItem{
		Id:               "MK",
		SortKey:          BuildMarketplaceDynamoId(item.Kind, item.Date, item.AwayTeam, item.HomeTeam),
		Spread:           item.Spread,
		Ttl:              item.Date.AddDate(0, 0, 1).Unix(),
		HomeAmount:       item.HomeAmount,
		Week:             item.Week,
		AwayAmount:       item.AwayAmount,
		AwayAbbreviation: item.AwayAbbreviation,
		HomeAbbreviation: item.HomeAbbreviation,
		AwayRecord:       item.AwayRecord,
		HomeRecord:       item.HomeRecord,
		EventId:          item.Id,
	}
}

func (item MarketplaceDynamoDbItem) GetItem() database.Item {
	expression := regexp.MustCompile(`(?P<Kind>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)`)

	match := expression.FindStringSubmatch(item.SortKey)

	paramsMap := make(map[string]string)
	for i, name := range expression.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	date, _ := time.Parse(time.RFC3339, paramsMap["Date"])

	return MarketplaceItem{
		AwayTeam:         paramsMap["AwayTeam"],
		HomeTeam:         paramsMap["HomeTeam"],
		Date:             date,
		Kind:             paramsMap["Kind"],
		Spread:           item.Spread,
		HomeAmount:       item.HomeAmount,
		AwayAmount:       item.AwayAmount,
		AwayAbbreviation: item.AwayAbbreviation,
		HomeAbbreviation: item.HomeAbbreviation,
		AwayRecord:       item.AwayRecord,
		HomeRecord:       item.HomeRecord,
		Id:               item.EventId,
		Week:             item.Week,
	}
}

func BuildMarketplaceDynamoId(kind string, date time.Time, awayTeam string, homeTeam string) string {
	return fmt.Sprintf("%s|%s|%s|%s", kind, date.Format(time.RFC3339), awayTeam, homeTeam)
}

func (s *MarketplaceService) GetItems(ctx context.Context) []MarketplaceItem {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "MK"},
		},
	})
}

func (s *MarketplaceService) ModifyAmount(ctx context.Context, bid bid.Bid) {

	var updateExpression *string

	if bid.ChosenCompetitor == bid.AwayTeam {
		updateExpression = aws.String("add awayAmount :amount")
	} else {
		updateExpression = aws.String("add homeAmount :amount")
	}

	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: strconv.FormatInt(bid.Amount, 10)},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: "MK"},
			"sortKey": &types.AttributeValueMemberS{Value: BuildMarketplaceDynamoId(bid.Kind, bid.Date, bid.AwayTeam, bid.HomeTeam)},
		},
		UpdateExpression: updateExpression,
	}

	s.databaseService.UpdateItem(ctx, input)

}

func (s *MarketplaceService) Write(ctx context.Context, items []MarketplaceItem) {
	s.databaseService.Write(ctx, items)
}
