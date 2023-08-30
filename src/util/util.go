package util

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type DefaultResponse struct {
	Message string `json:"message"`
}

func ApigatewayResponse(body string, statusCode int) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{Body: body, Headers: map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "X-Amz-Date,X-Api-Key,X-Amz-Security-Token,X-Requested-With,X-Auth-Token,Referer,User-Agent,Origin,Content-Type,Authorization,Accept,Access-Control-Allow-Methods,Access-Control-Allow-Origin,Access-Control-Allow-Headers",
	},
		StatusCode: statusCode}, nil
}

func GetAwsConfig(ctx context.Context) aws.Config {
	cfg, _ := config.LoadDefaultConfig(ctx)
	return cfg
}
func GetDynamoClient(config aws.Config) *dynamodb.Client {
	client := dynamodb.NewFromConfig(config)

	return client
}

func Filter[T any](slice []T, test func(T) bool) []T {
	ret := make([]T, 0, len(slice))
	for _, element := range slice {
		if test(element) {
			ret = append(ret, element)
		}
	}
	return ret
}

type LockItem struct {
	Id      string `dynamodbav:"id"`
	SortKey string `dynamodbav:"sortKey"`
	Ttl     int64  `dynamodbav:"ttl"`
}

func GetLock(ctx context.Context, key string, client *dynamodb.Client) error {
	av, _ := attributevalue.MarshalMap(LockItem{
		Id:      "LOCK",
		SortKey: key,
		Ttl:     time.Now().Add(time.Duration(6e+10)).Unix(),
	})

	input := &dynamodb.PutItemInput{
		Item:                av,
		TableName:           aws.String("BackendStack-table8235A42E-1GC3LE21GNUV8"),
		ConditionExpression: aws.String("attribute_not_exists(sortKey)"),
	}

	_, err := client.PutItem(ctx, input)

	if err != nil {
		return err
	}
	return nil
}

func Query[T interface{}](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) []T {

	dynamoItems := executeQuery[T](ctx, client, input)

	return dynamoItems
}

func executeQuery[T interface{}](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) []T {
	resp, _ := client.Query(ctx, input)

	var items []T

	attributevalue.UnmarshalListOfMaps(resp.Items, &items)

	if len(resp.LastEvaluatedKey) > 0 {
		input.ExclusiveStartKey = resp.LastEvaluatedKey
		return append(items, executeQuery[T](ctx, client, input)...)
	}

	return items
}
