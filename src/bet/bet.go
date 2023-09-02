package bet

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/database"
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

type Service interface {
	GetBetsByEventDate(ctx context.Context, date string) []Bet
	GetBetsByDate(ctx context.Context, date string) []Bet
	GetBetsByUser(ctx context.Context, user string) []Bet
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

func BuildId(bet Bet) string {
	return fmt.Sprintf("BET|%s", bet.Date.Format(format))
}

func (item Bet) GetDynamoItem() database.DynamoItem {
	//@TODO seems hacky
	return BetDynamoItem{
		Id:           BuildId(item),
		SortKey:      BuildSortKey(item),
		Amount:       item.Amount,
		Gsi1_id:      fmt.Sprintf("BET|%s", item.AwayUser),
		Gsi1_sortKey: item.Date.Format(time.RFC3339),
		Gsi2_id:      fmt.Sprintf("BET|%s", item.HomeUser),
		Gsi2_sortKey: item.Date.Format(time.RFC3339),
		Gsi3_id:      "BET",
		Gsi3_sortKey: item.Date.Format(time.RFC3339),
		Status:       item.Status,
		Spread:       item.Spread,
		Ttl:          item.Date.AddDate(0, 0, 1).Unix(),
	}
}

const format = "20060102"

func BuildSortKey(bet Bet) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", bet.Kind, bet.AwayTeam, bet.HomeTeam, bet.AwayUser, bet.HomeUser)
}

func (s *BetService) GetBetsByDate(ctx context.Context, date string) []Bet {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("gsi3_id = :id"),
		IndexName:              aws.String("gsi3"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: "BET"},
		},
	})
}

func (s *BetService) GetBetsByEventDate(ctx context.Context, date string) []Bet {
	return s.databaseService.Query(ctx, &dynamodb.QueryInput{
		TableName:              aws.String(os.Getenv("TABLE_NAME")),
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":id": &types.AttributeValueMemberS{Value: fmt.Sprintf("BET|%s", date)},
		},
	})
}

func (s *BetService) GetBetsByUser(ctx context.Context, user string) []Bet {
	betChannel := make(chan []Bet)

	for i := 1; i < 3; i++ {
		go func(num int) {
			resp := s.databaseService.Query(ctx, &dynamodb.QueryInput{
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
	retSlice := make([]Bet, 0)
	for i := 1; i < 3; i++ {
		respSlice := <-betChannel
		retSlice = append(retSlice, respSlice...)
	}

	return retSlice
}

func (bet BetDynamoItem) GetItem() database.Item {
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

func (s *BetService) Write(ctx context.Context, items []Bet) {
	s.databaseService.Write(ctx, items)
}
