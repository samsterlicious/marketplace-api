package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"samster.link/marketplace"
	"samster.link/util"
)

func isRecent(item marketplace.MarketplaceItem) bool {
	return item.Date.After(time.Now())
}

func handleCreate(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	marketplaceEvents := marketplace.ConvertMarketplaceItems(marketplace.GetMarketplaceItems(os.Getenv("TABLE_NAME")))

	resp, _ := json.Marshal(util.Filter(marketplaceEvents, isRecent))

	return util.ApigatewayResponse(string(resp))
}

func main() {
	lambda.Start(handleCreate)
}
