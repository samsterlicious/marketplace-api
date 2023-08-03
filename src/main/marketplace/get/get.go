package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"samster.link/marketplace"
)

func handleCreate(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	marketplaceEvents := marketplace.ConvertMarketplaceItems(marketplace.GetMarketplaceItems())

	resp, _ := json.Marshal(&marketplaceEvents)

	return events.APIGatewayProxyResponse{Body: string(resp), StatusCode: 200}, nil
}

func main() {
	lambda.Start(handleCreate)
}
