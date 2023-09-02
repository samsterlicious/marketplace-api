package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/database"
	"sammy.link/user"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, userService user.Service) (events.APIGatewayV2HTTPResponse, error) {

	email := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]
	resp, err := userService.GetUser(ctx, email)

	if err != nil {
		newUser := user.Item{
			Email: email,
			Total: 0,
		}
		userService.Create(ctx, newUser)

		resp = newUser
	}

	jsonUser, _ := json.Marshal(resp)

	return util.ApigatewayResponse(string(jsonUser), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleGet(ctx, request, user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)))
		})
}
