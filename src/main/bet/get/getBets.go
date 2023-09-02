package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bet"
	"sammy.link/database"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, betService bet.Service) (events.APIGatewayV2HTTPResponse, error) {

	var bets []bet.Bet
	if date, ok := request.PathParameters["date"]; ok {
		bets = betService.GetBetsByDate(ctx, date)

	} else {
		bets = betService.GetBetsByUser(ctx, request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"])
	}

	jsonBets, _ := json.Marshal(bets)
	return util.ApigatewayResponse(string(jsonBets), 200)
}

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handleGet(ctx, request, bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)))
	})
}
