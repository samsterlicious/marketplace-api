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

	users := userService.GetUsers(ctx)

	jsonBids, _ := json.Marshal(users)

	return util.ApigatewayResponse(string(jsonBids), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleGet(ctx, request, user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)))
		})
}
