package outcome

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/database"
)

type OutcomeDynamoItem struct {
	Id           string `dynamodbav:"id"`
	SortKey      string `dynamodbav:"sortKey"`
	Gsi1_id      string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey string `dynamodbav:"gsi1_sortKey"`
	Gsi2_id      string `dynamodbav:"gsi2_id"`
	Gsi2_sortKey string `dynamodbav:"gsi2_sortKey"`
}

type OutcomeItem struct {
	Winner  string `json:"winner"`
	Loser   string `json:"loser"`
	EventId string `json:"eventId"`
	Week    string `json:"week"`
	Amount  int64  `json:"amount"`
	Id      string `json:"id"`
}

type Service interface {
	GetByUser(ctx context.Context, user string) []OutcomeItem
	Write(ctx context.Context, outcomes []OutcomeItem)
}
type OutcomeService struct {
	databaseService database.Service[OutcomeDynamoItem, OutcomeItem]
}

func NewService(databaseService database.Service[OutcomeDynamoItem, OutcomeItem]) Service {
	return &OutcomeService{
		databaseService: databaseService,
	}
}

func (item OutcomeItem) getDynamoId() string {
	return fmt.Sprintf("OUT|%s", item.Week)
}

func (dynamoItem OutcomeDynamoItem) GetItem() database.Item {
	amount, _ := strconv.ParseInt(dynamoItem.Gsi1_sortKey, 10, 64)
	return OutcomeItem{
		Winner:  dynamoItem.Gsi1_id,
		Loser:   dynamoItem.Gsi2_id,
		EventId: dynamoItem.Gsi2_sortKey,
		Week:    strings.Split(dynamoItem.Id, "|")[1],
		Amount:  amount,
	}
}

func (item OutcomeItem) GetDynamoItem() database.DynamoItem {
	return OutcomeDynamoItem{
		Id:           item.getDynamoId(),
		SortKey:      item.Id,
		Gsi1_id:      item.Winner,
		Gsi1_sortKey: fmt.Sprintf("%d", item.Amount),
		Gsi2_id:      item.Loser,
		Gsi2_sortKey: item.EventId,
	}
}

func (s *OutcomeService) GetByUser(ctx context.Context, user string) []OutcomeItem {
	var sliceOne, sliceTwo []OutcomeItem
	sliceChan := make(chan []OutcomeItem, 2)
	for i := 1; i < 3; i++ {
		go func(sliceChan chan []OutcomeItem, num int) {
			sliceChan <- s.databaseService.Query(ctx, &dynamodb.QueryInput{
				TableName:              aws.String(os.Getenv("TABLE_NAME")),
				KeyConditionExpression: aws.String(fmt.Sprintf("gsi%d_id = :id", num)),
				IndexName:              aws.String(fmt.Sprintf("gsi%d", num)),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":id": &types.AttributeValueMemberS{Value: user},
				},
			})

		}(sliceChan, i)
	}

	sliceOne = <-sliceChan
	sliceTwo = <-sliceChan

	return append(sliceOne, sliceTwo...)
}

func (s *OutcomeService) Write(ctx context.Context, outcomes []OutcomeItem) {
	s.databaseService.Write(ctx, outcomes)
}
