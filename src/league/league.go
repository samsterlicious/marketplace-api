package league

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
	"sammy.link/database"
)

type UserInLeagueDynamoItem struct {
	Id      string `dynamodbav:"id"`
	SortKey string `dynamodbav:"sortKey"`
	Name    string `dynamodbav:"na"`
	Total   int64  `dynamodbav:"amount"`
}

type UserInLeagueItem struct {
	Email  string `json:"email"`
	Name   string `json:"name"`
	League string `json:"league"`
	Total  int64  `json:"total"`
}

type LeagueDynamoItem struct {
	Id        string `dynamodbav:"id"`
	SortKey   string `dynamodbav:"sortKey"`
	AdminUser string `dynamodbav:"name"`
}

type LeagueItem struct {
	Name      string `json:"name"`
	AdminUser string `json:"adminUser"`
}

type Service interface {
	AddUser(ctx context.Context, email UserInLeagueItem)
	GetUsers(ctx context.Context, league string) []UserInLeagueItem
	Create(ctx context.Context, league LeagueItem)
	UpdateUserName(ctx context.Context, league string, email string, name string)
	UpdateUserAmount(ctx context.Context, league string, email string, amount int64)
}

func NewService(leagueDatabaseService database.Service[LeagueDynamoItem, LeagueItem], userDatabaseService database.Service[UserInLeagueDynamoItem, UserInLeagueItem]) Service {
	return &LeagueService{
		leagueDatabaseService: leagueDatabaseService,
		userDatabaseService:   userDatabaseService,
	}
}

type LeagueService struct {
	leagueDatabaseService database.Service[LeagueDynamoItem, LeagueItem]
	userDatabaseService   database.Service[UserInLeagueDynamoItem, UserInLeagueItem]
}

func (dynamoItem LeagueDynamoItem) GetItem() database.Item {
	return LeagueItem{
		Name:      dynamoItem.SortKey,
		AdminUser: dynamoItem.AdminUser,
	}
}

func (item LeagueItem) GetDynamoItem() database.DynamoItem {
	return LeagueDynamoItem{
		Id:        "LEAGUE",
		SortKey:   item.Name,
		AdminUser: item.AdminUser,
	}
}

func (dynamoItem UserInLeagueDynamoItem) GetItem() database.Item {
	return UserInLeagueItem{
		Email:  dynamoItem.SortKey,
		Name:   dynamoItem.Name,
		League: strings.Split(dynamoItem.Id, "|")[1],
		Total:  dynamoItem.Total,
	}
}

func (item UserInLeagueItem) GetDynamoItem() database.DynamoItem {
	return UserInLeagueDynamoItem{
		Id:      getUserId(item.League),
		SortKey: item.Email,
		Name:    item.Name,
		Total:   item.Total,
	}
}

func getUserId(league string) string {
	return fmt.Sprintf("L|%s", league)
}

func (s *LeagueService) UpdateUserName(ctx context.Context, league string, email string, name string) {
	s.userDatabaseService.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: getUserId(league)},
			"sortKey": &types.AttributeValueMemberS{Value: email},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":name": &types.AttributeValueMemberS{Value: name},
		},
		TableName:        aws.String(os.Getenv("TABLE_NAME")),
		UpdateExpression: aws.String("SET na = :name"),
	})
}

func (s *LeagueService) UpdateUserAmount(ctx context.Context, league string, email string, amount int64) {
	fmt.Printf("amount is %s; id = %s; sortKey =%s\n", strconv.FormatInt(amount, 10), getUserId(league), email)
	s.userDatabaseService.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: getUserId(league)},
			"sortKey": &types.AttributeValueMemberS{Value: email},
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":amount": &types.AttributeValueMemberN{Value: strconv.FormatInt(amount, 10)},
		},
		TableName:        aws.String(os.Getenv("TABLE_NAME")),
		UpdateExpression: aws.String("add amount :amount"),
	})
}

func (s *LeagueService) AddUser(ctx context.Context, item UserInLeagueItem) {
	s.userDatabaseService.Write(ctx, []UserInLeagueItem{item})
}

func (s *LeagueService) GetUsers(ctx context.Context, league string) []UserInLeagueItem {
	return s.userDatabaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("L|%s", league)},
		},
	})

}

func (s *LeagueService) Create(ctx context.Context, league LeagueItem) {
	s.leagueDatabaseService.Write(ctx, []LeagueItem{league})
}
