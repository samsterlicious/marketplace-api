package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"sammy.link/util"
)

type Item interface {
	GetDynamoItem() DynamoItem
}

type DynamoItem interface {
	GetItem() Item
}

type Service[D DynamoItem, I Item] interface {
	BatchWriteItem(ctx context.Context, input *dynamodb.BatchWriteItemInput)
	Get(ctx context.Context, params *dynamodb.GetItemInput) (I, error)
	Write(ctx context.Context, items []I)
	Query(ctx context.Context, params *dynamodb.QueryInput) []I
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput)
	Lock(ctx context.Context, key string)
	ReleaseLock(ctx context.Context, key string)
	Delete(ctx context.Context, input *dynamodb.DeleteItemInput)
}

type DynamoDbService[D DynamoItem, I Item] struct {
	client *dynamodb.Client
}

var client *dynamodb.Client

func GetDatabaseService[D DynamoItem, I Item](ctx context.Context) *DynamoDbService[D, I] {
	if client == nil {
		defaultConfig, err := util.GetAwsConfig(ctx)

		if err != nil {
			fmt.Println(err.Error())
		}

		client = dynamodb.NewFromConfig(defaultConfig)
	}
	return &DynamoDbService[D, I]{
		client: client,
	}
}

func (s *DynamoDbService[D, I]) Get(ctx context.Context, params *dynamodb.GetItemInput) (I, error) {
	var getItem I

	resp, err := s.client.GetItem(ctx, params)

	if err != nil {
		fmt.Println(err.Error())
		return getItem, err
	}

	err = attributevalue.UnmarshalMap(resp.Item, getItem)

	if err != nil {
		fmt.Println(err.Error())
		return getItem, err
	}

	return getItem, nil
}

func (s *DynamoDbService[D, I]) Delete(ctx context.Context, input *dynamodb.DeleteItemInput) {
	s.client.DeleteItem(ctx, input)
}

func (s *DynamoDbService[D, I]) BatchWriteItem(ctx context.Context, input *dynamodb.BatchWriteItemInput) {
	s.client.BatchWriteItem(ctx, input)
}

func (s *DynamoDbService[D, I]) Write(ctx context.Context, items []I) {
	requestItems := map[string][]types.WriteRequest{}

	putItems := make([]types.WriteRequest, len(items))

	for i, item := range items {
		attributeValueMap, _ := attributevalue.MarshalMap(item.GetDynamoItem())

		putItems[i] = types.WriteRequest{PutRequest: &types.PutRequest{
			Item: attributeValueMap,
		}}
	}

	requestItems[os.Getenv("TABLE_NAME")] = putItems

	_, err := s.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
		RequestItems: requestItems,
	},
	)

	if err != nil {
		fmt.Println(err.Error())
	}
}

func (s *DynamoDbService[D, I]) Query(ctx context.Context, params *dynamodb.QueryInput) []I {
	dynamoItems := Query[D](ctx, s.client, params)

	itemSlice := make([]I, len(dynamoItems))

	for i, dynamoItem := range dynamoItems {
		item := dynamoItem.GetItem().(I)
		itemSlice[i] = item
	}

	return itemSlice
}

func (s *DynamoDbService[D, I]) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput) {
	_, err := s.client.UpdateItem(ctx, params)

	if err != nil {
		fmt.Println(err.Error())
	}
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
		TableName:           aws.String(os.Getenv("TABLE_NAME")),
		ConditionExpression: aws.String("attribute_not_exists(sortKey)"),
	}

	_, err := client.PutItem(ctx, input)

	if err != nil {
		return err
	}
	return nil
}

func (s *DynamoDbService[D, I]) Lock(ctx context.Context, key string) {
	for {
		err := GetLock(ctx, key, s.client)
		if err != nil {
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}
}

func (s *DynamoDbService[D, I]) ReleaseLock(ctx context.Context, key string) {
	s.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		Key: map[string]types.AttributeValue{
			"id":      &types.AttributeValueMemberS{Value: "LOCK"},
			"sortKey": &types.AttributeValueMemberS{Value: key},
		},
		TableName: aws.String(os.Getenv("TABLE_NAME")),
	})
}

func Query[D DynamoItem](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) []D {

	dynamoItems := executeQuery[D](ctx, client, input)

	return dynamoItems
}

func executeQuery[D DynamoItem](ctx context.Context, client *dynamodb.Client, input *dynamodb.QueryInput) []D {
	resp, err := client.Query(ctx, input)

	if err != nil {
		fmt.Println(err.Error())
	}

	var items []D

	attributevalue.UnmarshalListOfMaps(resp.Items, &items)

	if len(resp.LastEvaluatedKey) > 0 {
		input.ExclusiveStartKey = resp.LastEvaluatedKey
		return append(items, executeQuery[D](ctx, client, input)...)
	}

	return items
}
