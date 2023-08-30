package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"sammy.link/bet"
	"sammy.link/util"
)

var dynamoClient *dynamodb.Client

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	if dynamoClient == nil {
		dynamoClient = util.GetDynamoClient(util.GetAwsConfig(ctx))
	}

	var dynamoBets []bet.BetDynamoItem
	if date, ok := request.PathParameters["date"]; ok {
		dynamoBets = bet.GetBetsByDate(ctx, date, dynamoClient)

	} else {
		dynamoBets = bet.GetBetsByUser(ctx, request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"], dynamoClient)
	}

	bids := make([]bet.Bet, 0, len(dynamoBets))

	for _, bet := range dynamoBets {
		bids = append(bids, bet.GetItem())
	}

	jsonBids, _ := json.Marshal(bids)
	return util.ApigatewayResponse(string(jsonBids), 200)
}

func main() {
	lambda.Start(handleGet)
}
