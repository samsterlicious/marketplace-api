package outcome

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoItem struct {
	Id           string `dynamodbav:"id"`
	SortKey      string `dynamodbav:"sortKey"`
	Gsi1_id      string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey string `dynamodbav:"gsi1_sortKey"`
	Gsi2_id      string `dynamodbav:"gsi2_id"`
	Gsi2_sortKey string `dynamodbav:"gsi2_sortKey"`
}

type Outcome struct {
	Winner  string `json:"winner"`
	Loser   string `json:"loser"`
	EventId string `json:"eventId"`
	Week    string `json:"week"`
	Amount  int64  `json:"winnerAmount"`
	Id      string `json:"id"`
}

func GetDynamoId(week string) string {
	return fmt.Sprintf("OUT|%s", week)
}

func GetDynamoSortKey(guid string) string {
	return fmt.Sprintf("%s", guid)
}

func (outcome Outcome) getDynamoItem() DynamoItem {
	return DynamoItem{
		Id:           GetDynamoId(outcome.Week),
		SortKey:      GetDynamoSortKey(outcome.Id),
		Gsi1_id:      outcome.Winner,
		Gsi1_sortKey: fmt.Sprintf("%d", outcome.Amount),
		Gsi2_id:      outcome.Loser,
		Gsi2_sortKey: outcome.EventId,
	}
}

func Write(ctx context.Context, outcomes []Outcome, client *dynamodb.Client) {

	requestItems := map[string][]types.WriteRequest{}

	putItems := make([]types.WriteRequest, len(outcomes))

	for i, item := range outcomes {
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
