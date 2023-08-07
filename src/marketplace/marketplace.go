package marketplace

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type MarketplaceDynamoDbItem struct {
	Id       string `dynamodbav:"id"`
	SortKey  string `dynamodbav:"sortKey"`
	Spread   string `dynamodbav:"spread"`
	Name     string `dynamodbav:"name"`
	AwayLogo string `dynamodbav:"awayLogo"`
	HomeLogo string `dynamodbav:"homeLogo"`
	Ttl      int64  `dynamodbav:"ttl"`
}

func (item MarketplaceDynamoDbItem) String() string {
	return fmt.Sprintf("id:%s\tsortKey:%s", item.Id, item.SortKey)
}

type MarketplaceItem struct {
	Name           string    `json:"name"`
	AwayCompetitor string    `json:"awayCompetitor"`
	HomeCompetitor string    `json:"homeCompetitor"`
	Date           time.Time `json:"date"`
	Sport          string    `json:"sport"`
	League         string    `json:"league"`
	AwayLogo       string    `json:"awayLogo"`
	HomeLogo       string    `json:"homeLogo"`
	Spread         string    `json:"spread"`
}

func getDynamoClient() *dynamodb.DynamoDB {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	svc := dynamodb.New(sess)

	return svc
}

func ConvertMarketplaceItems(items []MarketplaceDynamoDbItem) []MarketplaceItem {
	response := make([]MarketplaceItem, 0, len(items))

	for _, item := range items {
		response = append(response, item.GetMarketplaceItem())
	}
	return response
}

func (item MarketplaceDynamoDbItem) GetMarketplaceItem() MarketplaceItem {
	expression := regexp.MustCompile(`(?P<Sport>[^|]+)\|(?P<League>[^|]+)\|(?P<Date>[^|]+)\|(?P<AwayCompetitor>[^|]+)\|(?P<HomeCompetitor>[^|]+)`)

	match := expression.FindStringSubmatch(item.SortKey)

	paramsMap := make(map[string]string)
	for i, name := range expression.SubexpNames() {
		if i > 0 && i <= len(match) {
			paramsMap[name] = match[i]
		}
	}

	date, _ := time.Parse(time.RFC3339, paramsMap["Date"])

	return MarketplaceItem{
		Name:           item.Name,
		AwayCompetitor: paramsMap["AwayCompetitor"],
		HomeCompetitor: paramsMap["HomeCompetitor"],
		Date:           date,
		Sport:          paramsMap["Sport"],
		League:         paramsMap["League"],
		AwayLogo:       item.AwayLogo,
		HomeLogo:       item.HomeLogo,
		Spread:         item.Spread,
	}
}

func BuildMarketplaceDynamoId(sport string, league string, date time.Time, awayTeam string, homeTeam string) string {
	return fmt.Sprintf("%s|%s|%s|%s|%s", sport, league, date.Format(time.RFC3339), awayTeam, homeTeam)
}

func GetMarketplaceItems(tableName string) []MarketplaceDynamoDbItem {
	svc := getDynamoClient()
	resp, err := svc.Query(&dynamodb.QueryInput{
		TableName:              &tableName,
		KeyConditionExpression: aws.String("id = :id"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":id": {
				S: aws.String("MK"),
			},
		},
	})

	if err != nil {
		fmt.Println(err.Error())
	}

	var marketplaceItems []MarketplaceDynamoDbItem

	err = dynamodbattribute.UnmarshalListOfMaps(resp.Items, &marketplaceItems)

	if err != nil {
		fmt.Println(err.Error())
	}

	return marketplaceItems
}

func WriteMarketplaceItems(items []MarketplaceItem, tableName string, waitGroup *sync.WaitGroup) {

	svc := getDynamoClient()

	defer waitGroup.Done()

	requestItems := map[string][]*dynamodb.WriteRequest{}

	putItems := make([]*dynamodb.WriteRequest, len(items))

	for i, item := range items {
		attributeValueMap, _ := dynamodbattribute.MarshalMap(MarketplaceDynamoDbItem{
			Id:       "MK",
			SortKey:  BuildMarketplaceDynamoId(item.Sport, item.League, item.Date, item.AwayCompetitor, item.HomeCompetitor),
			AwayLogo: item.AwayLogo,
			HomeLogo: item.HomeLogo,
			Spread:   item.Spread,
			Name:     item.Name,
			Ttl:      item.Date.AddDate(0, 0, 1).Unix(),
		})

		putItems[i] = &dynamodb.WriteRequest{PutRequest: &dynamodb.PutRequest{
			Item: attributeValueMap,
		}}
	}

	requestItems[tableName] = putItems

	_, err := svc.BatchWriteItem(&dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	},
	)

	if err != nil {
		fmt.Println(err.Error())
	}

}
