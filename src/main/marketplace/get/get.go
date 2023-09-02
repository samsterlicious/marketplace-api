package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"sammy.link/database"
	"sammy.link/marketplace"
	"sammy.link/util"
)

func isRecent(item marketplace.MarketplaceItem) bool {
	return item.Date.After(time.Now())
}

func handleCreate(ctx context.Context, request events.APIGatewayV2HTTPRequest, service marketplace.Service) (events.APIGatewayV2HTTPResponse, error) {

	marketplaceEvents := service.GetItems(ctx)

	resp, _ := json.Marshal(util.Filter(marketplaceEvents, isRecent))

	return util.ApigatewayResponse(string(resp), 200)
}

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handleCreate(ctx, request, marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
	})
}
