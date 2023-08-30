package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"sammy.link/marketplace"
	"sammy.link/util"
)

func isRecent(item marketplace.MarketplaceItem) bool {
	return item.Date.After(time.Now())
}

var dynamoClient *dynamodb.Client

func handleCreate(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if dynamoClient == nil {
		dynamoClient = util.GetDynamoClient(util.GetAwsConfig(ctx))
	}

	marketplaceEvents := marketplace.ConvertMarketplaceItems(marketplace.GetMarketplaceItems(ctx, dynamoClient))

	resp, _ := json.Marshal(util.Filter(marketplaceEvents, isRecent))

	return util.ApigatewayResponse(string(resp), 200)
}

func main() {
	lambda.Start(handleCreate)
}
