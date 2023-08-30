package marketplace

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
	"sammy.link/bid"
	"sammy.link/util"
)

type MarketplaceDynamoDbItem struct {
	Id         string `dynamodbav:"id"`
	SortKey    string `dynamodbav:"sortKey"`
	Spread     string `dynamodbav:"spread"`
	Ttl        int64  `dynamodbav:"ttl"`
	HomeAmount int64  `dynamodbav:"homeAmount"`
	AwayAmount int64  `dynamodbav:"awayAmount"`
}

func (item MarketplaceDynamoDbItem) String() string {
	return fmt.Sprintf("id:%s\tsortKey:%s", item.Id, item.SortKey)
}

type MarketplaceItem struct {
	AwayTeam   string    `json:"awayTeam"`
	HomeTeam   string    `json:"homeTeam"`
	Date       time.Time `json:"date"`
	Kind       string    `json:"kind"`
	Spread     string    `json:"spread"`
	HomeAmount int64     `json:"homeAmount"`
	AwayAmount int64     `json:"awayAmount"`
}

func ConvertMarketplaceItems(items []MarketplaceDynamoDbItem) []MarketplaceItem {
	response := make([]MarketplaceItem, 0, len(items))

	for _, item := range items {
		response = append(response, item.GetMarketplaceItem())
	}
	return response
}

func (item MarketplaceDynamoDbItem) GetMarketplaceItem() MarketplaceItem {
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
		AwayTeam:   paramsMap["AwayTeam"],
		HomeTeam:   paramsMap["HomeTeam"],
		Date:       date,
		Kind:       paramsMap["Kind"],
		Spread:     item.Spread,
		HomeAmount: item.HomeAmount,
		AwayAmount: item.AwayAmount,
	}
}

func BuildMarketplaceDynamoId(kind string, date time.Time, awayTeam string, homeTeam string) string {
	return fmt.Sprintf("%s|%s|%s|%s", kind, date.Format(time.RFC3339), awayTeam, homeTeam)
}

func GetMarketplaceItems(ctx context.Context, client *dynamodb.Client) []MarketplaceDynamoDbItem {
	return util.Query[MarketplaceDynamoDbItem](ctx, client, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "MK"},
		},
	})
}

func ModifyAmount(ctx context.Context, bid bid.Bid, client *dynamodb.Client) {

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

	client.UpdateItem(ctx, input)

}

func WriteMarketplaceItems(ctx context.Context, items []MarketplaceItem, waitGroup *sync.WaitGroup, client *dynamodb.Client) {

	defer waitGroup.Done()

	requestItems := map[string][]types.WriteRequest{}

	putItems := make([]types.WriteRequest, len(items))

	for i, item := range items {
		attributeValueMap, _ := attributevalue.MarshalMap(MarketplaceDynamoDbItem{
			Id:         "MK",
			SortKey:    BuildMarketplaceDynamoId(item.Kind, item.Date, item.AwayTeam, item.HomeTeam),
			Spread:     item.Spread,
			Ttl:        item.Date.AddDate(0, 0, 1).Unix(),
			HomeAmount: 0,
			AwayAmount: 0,
		})

		putItems[i] = types.WriteRequest{PutRequest: &types.PutRequest{
			Item: attributeValueMap,
		}}
	}

	requestItems[os.Getenv("TABLE_NAME")] = putItems

	_, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	})

	if err != nil {
		fmt.Println(err.Error())
	}

}
