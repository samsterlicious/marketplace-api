package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, bidService bid.Service) (events.APIGatewayV2HTTPResponse, error) {

	bids := bidService.GetBidsByUser(ctx, request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"])

	jsonBids, _ := json.Marshal(bids)

	return util.ApigatewayResponse(string(jsonBids), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleGet(ctx, request, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)))
		})
}
