package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"sammy.link/database"
	"sammy.link/league"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, service league.Service) (events.APIGatewayV2HTTPResponse, error) {

	l := request.PathParameters["league"]

	users := service.GetUsers(ctx, l)

	resp, _ := json.Marshal(users)
	return util.ApigatewayResponse(string(resp), 200)
}

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handleGet(ctx, request, league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
			database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
	})
}
