package user

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"sammy.link/database"
)

type DynamoItem struct {
	Id      string `dynamodbav:"id"`
	SortKey string `dynamodbav:"sortKey"`
	Total   int64  `dynamodbav:"amount"`
}

type Item struct {
	Email string `json:"email"`
	Total int64  `json:"total"`
}

type Service interface {
	GetUser(ctx context.Context, user string) (Item, error)
	GetUsers(ctx context.Context) []Item
	Create(ctx context.Context, item Item)
	Update(ctx context.Context, user string, amount int64)
}

type UserService struct {
	databaseService database.Service[DynamoItem, Item]
}

func NewService(databaseService database.Service[DynamoItem, Item]) Service {
	return &UserService{
		databaseService: databaseService,
	}
}

func (dynamoItem DynamoItem) GetItem() database.Item {
	return Item{
		Email: dynamoItem.SortKey,
		Total: dynamoItem.Total,
	}
}

func (item Item) GetDynamoItem() database.DynamoItem {
	return DynamoItem{
		Id:      "USER",
		SortKey: item.Email,
		Total:   item.Total,
	}
}

func (s *UserService) GetUser(ctx context.Context, user string) (Item, error) {
	return s.databaseService.Get(ctx, &dynamodb.GetItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: "USER"},
			"sortKey": &types.AttributeValueMemberS{Value: user},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
}

func (s *UserService) GetUsers(ctx context.Context) []Item {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "USER"},
		},
	})
}

func (s *UserService) Create(ctx context.Context, item Item) {
	s.databaseService.Write(ctx, []Item{item})
}

func (s *UserService) Update(ctx context.Context, user string, amount int64) {
	input := &dynamodb.UpdateItemInput{
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", amount)},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: "USER"},
			"sortKey": &types.AttributeValueMemberS{Value: user},
		},
		UpdateExpression: aws.String("add amount :amount"),
	}

	s.databaseService.UpdateItem(ctx, input)
}
