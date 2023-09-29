package bet

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
	"sammy.link/database"
)

type BetDynamoItem struct {
	Id               string `dynamodbav:"id"`
	SortKey          string `dynamodbav:"sortKey"`
	Gsi1_id          string `dynamodbav:"gsi1_id"`
	Gsi1_sortKey     string `dynamodbav:"gsi1_sortKey"`
	Gsi2_id          string `dynamodbav:"gsi2_id"`
	Gsi2_sortKey     string `dynamodbav:"gsi2_sortKey"`
	Gsi3_id          string `dynamodbav:"gsi3_id"`
	Gsi3_sortKey     string `dynamodbav:"gsi3_sortKey"`
	Amount           int64  `dynamodbav:"amount"`
	Status           string `dynamodbav:"status"`
	Spread           string `dynamodbav:"spread"`
	Ttl              int64  `dynamodbav:"ttl"`
	HomeAbbreviation string `dynamodbav:"ha"`
	AwayAbbreviation string `dynamodbav:"aa"`
}

type Bet struct {
	Div              string    `json:"div"`
	AwayUser         string    `json:"awayUser"`
	HomeUser         string    `json:"homeUser"`
	Amount           int64     `json:"amount"`
	AwayTeam         string    `json:"awayTeam"`
	HomeTeam         string    `json:"homeTeam"`
	Status           string    `json:"status"`
	Spread           string    `json:"spread"`
	Kind             string    `json:"kind"`
	Week             int       `json:"week"`
	CreateDate       time.Time `json:"createDate"`
	Date             time.Time `json:"date"`
	HomeAbbreviation string    `json:"homeAbbreviation"`
	AwayAbbreviation string    `json:"awayAbbreviation"`
}

type Service interface {
	GetBetsByEventDate(ctx context.Context, date string) []Bet
	GetBetsByWeek(ctx context.Context, div string, week string) []Bet
	GetBetsByUser(ctx context.Context, user string, isGsi2 bool) []Bet
	Write(ctx context.Context, items []Bet)
}

type BetService struct {
	databaseService database.Service[BetDynamoItem, Bet]
}

func NewService(databaseService database.Service[BetDynamoItem, Bet]) Service {
	return &BetService{
		databaseService: databaseService,
	}
}

func BuildId(div string, week string) string {
	return fmt.Sprintf("BET|%s|%s", div, week)
}

func (item Bet) GetDynamoItem() database.DynamoItem {
	//@TODO seems hacky
	return BetDynamoItem{
		Id:               BuildId(item.Div, strconv.Itoa(item.Week)),
		SortKey:          BuildSortKey(item),
		Amount:           item.Amount,
		Gsi1_id:          "BET",
		Gsi1_sortKey:     item.Date.Format(format),
		Gsi2_id:          fmt.Sprintf("BET|%s", item.AwayUser),
		Gsi2_sortKey:     item.Date.Format(time.RFC3339),
		Gsi3_id:          fmt.Sprintf("BET|%s", item.HomeUser),
		Gsi3_sortKey:     item.Date.Format(time.RFC3339),
		Status:           item.Status,
		Spread:           item.Spread,
		Ttl:              item.Date.AddDate(0, 0, 1).Unix(),
		HomeAbbreviation: item.HomeAbbreviation,
		AwayAbbreviation: item.AwayAbbreviation,
	}
}

const format = "20060102"

func BuildSortKey(bet Bet) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s", bet.Kind, bet.AwayTeam, bet.HomeTeam, bet.AwayUser, bet.HomeUser, bet.Date.Format(time.RFC3339))
}

func (s *BetService) GetBetsByWeek(ctx context.Context, div string, week string) []Bet {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: BuildId(div, week)},
		},
	})
}

func (s *BetService) GetBetsByUser(ctx context.Context, user string, isGsi2 bool) []Bet {
	if isGsi2 {
		return s.databaseService.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(os.Getenv("TABLE_NAME")),
			KeyConditionExpression: aws.String("gsi2_id = :id"),
			IndexName:              aws.String("gsi2"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", user)},
			},
		})
	} else {
		return s.databaseService.Query(ctx, &dynamodb.QueryInput{
			TableName:              aws.String(os.Getenv("TABLE_NAME")),
			KeyConditionExpression: aws.String("gsi3_id = :id"),
			IndexName:              aws.String("gsi3"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", user)},
			},
		})
	}
}

func (s *BetService) GetBetsByEventDate(ctx context.Context, date string) []Bet {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("gsi1_id = :id and gsi1_sortKey = :date"),
		IndexName:              aws.String("gsi1"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id":   &types.AttributeValueMemberS{Value: "BET"},
			":date": &types.AttributeValueMemberS{Value: date},
		},
	})
}

func (bet BetDynamoItem) GetItem() database.Item {
	idExpression := regexp.MustCompile(`BET\|(?P<Div>[^|]+)\|(?P<Week>[^|]+)`)
	// return fmt.Sprintf("%s|%s|%s|%s|%s", bet.Kind, bet.AwayTeam, bet.HomeTeam, bet.AwayUser, bet.HomeUser)
	sortKeyExpression := regexp.MustCompile(`(?P<Kind>[^|]+)\|(?P<AwayTeam>[^|]+)\|(?P<HomeTeam>[^|]+)\|(?P<AwayUser>[^|]+)\|(?P<HomeUser>[^|]+)\|(?P<CreateDate>[^|]+)`)

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
	week, _ := strconv.Atoi(paramsMap["Week"])
	return Bet{
		Kind:             paramsMap["Kind"],
		AwayTeam:         paramsMap["AwayTeam"],
		HomeTeam:         paramsMap["HomeTeam"],
		Spread:           bet.Spread,
		AwayUser:         paramsMap["AwayUser"],
		HomeUser:         paramsMap["HomeUser"],
		Status:           bet.Status,
		Amount:           bet.Amount,
		Week:             week,
		Div:              paramsMap["Div"],
		Date:             gameDate.AddDate(0, -1, -1),
		HomeAbbreviation: bet.HomeAbbreviation,
		AwayAbbreviation: bet.AwayAbbreviation,
	}
}

func (s *BetService) Write(ctx context.Context, items []Bet) {
	s.databaseService.Write(ctx, items)
}
