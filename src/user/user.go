package user

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"sammy.link/database"
)

type DynamoItem struct {
	Id      string `dynamodbav:"id"`
	SortKey string `dynamodbav:"sortKey"`
	Name    string `dynamodbav:"na"`
}

type Item struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	League string `json:"league"`
}

type Service interface {
	GetUser(ctx context.Context, user string) []Item
	Create(ctx context.Context, item Item)
	Update(ctx context.Context, user string, amount int64)
	UpdateName(ctx context.Context, item Item)
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
		Email:  strings.Split(dynamoItem.Id, "|")[1],
		Name:   dynamoItem.Name,
		League: dynamoItem.SortKey,
	}
}

func (item Item) GetDynamoItem() database.DynamoItem {
	return DynamoItem{
		Id:      getId(item.Email),
		SortKey: item.League,
		Name:    item.Name,
	}
}

func getId(email string) string {
	return fmt.Sprintf("U|%s", email)
}

func (s *UserService) UpdateName(ctx context.Context, item Item) {
	fmt.Printf("item = %+v", item)
	s.databaseService.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: getId(item.Email)},
			"sortKey": &types.AttributeValueMemberS{Value: item.League},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: item.Name},
		},
		TableName:        aws.String(os.Getenv("TABLE_NAME")),
		UpdateExpression: aws.String("SET na = :name"),
	})
}

func (s *UserService) GetUser(ctx context.Context, email string) []Item {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("U|%s", email)},
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
