package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/database"
	"sammy.link/league"
	"sammy.link/user"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, userService user.Service, leagueService league.Service) (events.APIGatewayV2HTTPResponse, error) {

	email := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]
	resp := userService.GetUser(ctx, email)

	if len(resp) < 1 {
		newUser := user.Item{
			Email:  email,
			League: "default",
			Name:   "",
		}
		userService.Create(ctx, newUser)
		leagueService.AddUser(ctx, league.UserInLeagueItem{
			Email:  email,
			Name:   "",
			League: "default",
			Total:  0,
		},
		)
		resp = []user.Item{newUser}
	}

	jsonUser, _ := json.Marshal(resp)

	return util.ApigatewayResponse(string(jsonUser), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleGet(ctx, request, user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)),
				league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
					database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
		})
}
