package bet

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/util"
)

type BetDynamoItem struct {
	Id           string `dynamodbav:"id"`
	SortKey      string `dynamodbav:"sortKey"`
	Gsi1_id      string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey string `dynamodbav:"gsi1_sortKey"`
	Gsi2_id      string `dynamodbav:"gsi2_id"`
	Gsi2_sortKey string `dynamodbav:"gsi2_sortKey"`
	Gsi3_id      string `dynamodbav:"gsi3_id"`
	Gsi3_sortKey string `dynamodbav:"gsi3_sortKey"`
	Amount       int64  `dynamodbav:"amount"`
	Status       string `dynamodbav:"status"`
	Spread       string `dynamodbav:"spread"`
	Ttl          int64  `dynamodbav:"ttl"`
}

type Bet struct {
	AwayUser string    `json:"awayUser"`
	HomeUser string    `json:"homeUser"`
	Amount   int64     `json:"amount"`
	AwayTeam string    `json:"awayTeam"`
	HomeTeam string    `json:"homeTeam"`
	Status   string    `json:"status"`
	Spread   string    `json:"spread"`
	Kind     string    `json:"kind"`
	Date     time.Time `json:"date"`
}

const format = "20060102"

func BuildId(bet Bet) string {
	return fmt.Sprintf("BET|%s", bet.Date.Format(format))
}

func BuildSortKey(bet Bet) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", bet.Kind, bet.AwayTeam, bet.HomeTeam, bet.AwayUser, bet.HomeUser)
}

func GetBetsByDate(ctx context.Context, date string, client *dynamodb.Client) []BetDynamoItem {
	return util.Query[BetDynamoItem](ctx, client, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("gsi3_id = :id"),
		IndexName:              aws.String("gsi3"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", date)},
		},
	})
}

func GetBetsByEventDate(ctx context.Context, date string, client *dynamodb.Client) []BetDynamoItem {
	return util.Query[BetDynamoItem](ctx, client, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", date)},
		},
	})
}

func GetBetsByUser(ctx context.Context, user string, client *dynamodb.Client) []BetDynamoItem {
	betChannel := make(chan []BetDynamoItem)

	for i := 1; i < 3; i++ {
		go func(num int) {
			resp := util.Query[BetDynamoItem](ctx, client, &dynamodb.QueryInput{
				TableName:              aws.String(os.Getenv("TABLE_NAME")),
				KeyConditionExpression: aws.String(fmt.Sprintf("gsi%d_id = :id", num)),
				IndexName:              aws.String(fmt.Sprintf("gsi%d", num)),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", user)},
				},
			})
			betChannel <- resp
		}(i)
	}
	retSlice := make([]BetDynamoItem, 0)
	for i := 1; i < 3; i++ {
		respSlice := <-betChannel
		retSlice = append(retSlice, respSlice...)
	}

	return retSlice
}

func (bet BetDynamoItem) GetItem() Bet {
	idExpression := regexp.MustCompile(`BET\|(?P<Day>[^|]+)`)
	// return fmt.Sprintf("%s|%s|%s|%s|%s", bet.Kind, bet.AwayTeam, bet.HomeTeam, bet.AwayUser, bet.HomeUser)
	sortKeyExpression := regexp.MustCompile(`(?P<Kind>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)\|(?P<AwayUser>[^|]+)\|(?P<HomeUser>[^|]+)`)

	idMatch := idExpression.FindStringSubmatch(bet.Id)
	sortKeyMatch := sortKeyExpression.FindStringSubmatch(bet.SortKey)

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

	gameDate := time.Unix(bet.Ttl, 0)

	return Bet{
		Kind:     paramsMap["Kind"],
		AwayTeam: paramsMap["AwayTeam"],
		HomeTeam: paramsMap["HomeTeam"],
		Spread:   bet.Spread,
		AwayUser: paramsMap["AwayUser"],
		HomeUser: paramsMap["HomeUser"],
		Status:   bet.Status,
		Amount:   bet.Amount,
		Date:     gameDate.AddDate(0, -1, -1),
	}
}

func ConvertDyanmoList(dynamoBids []BetDynamoItem) []Bet {
	bets := make([]Bet, 0, len(dynamoBids))

	for _, dynamoBid := range dynamoBids {
		bets = append(bets, dynamoBid.GetItem())
	}

	return bets
}

func WriteBets(ctx context.Context, items []Bet, client *dynamodb.Client, now time.Time) {

	requestItems := map[string][]types.WriteRequest{}

	putItems := make([]types.WriteRequest, len(items))

	for i, item := range items {
		attributeValueMap, _ := attributevalue.MarshalMap(BetDynamoItem{
			Id:           BuildId(item),
			SortKey:      BuildSortKey(item),
			Amount:       item.Amount,
			Gsi1_id:      fmt.Sprintf("BET|%s", item.AwayUser),
			Gsi1_sortKey: item.Date.Format(time.RFC3339),
			Gsi2_id:      fmt.Sprintf("BET|%s", item.HomeUser),
			Gsi2_sortKey: item.Date.Format(time.RFC3339),
			Gsi3_id:      fmt.Sprintf("BET|%s", now.Format(format)),
			Gsi3_sortKey: item.Date.Format(time.RFC3339),
			Status:       item.Status,
			Spread:       item.Spread,
			Ttl:          item.Date.AddDate(0, 1, 0).Unix(),
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
