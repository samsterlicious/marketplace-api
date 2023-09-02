package main

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/database"
	"sammy.link/outcome"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, outcomeService outcome.Service) (events.APIGatewayV2HTTPResponse, error) {

	outcomes := outcomeService.GetByUser(ctx, request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"])

	jsonOutcomes, _ := json.Marshal(outcomes)

	return util.ApigatewayResponse(string(jsonOutcomes), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleGet(ctx, request, outcome.NewService(database.GetDatabaseService[outcome.OutcomeDynamoItem, outcome.OutcomeItem](ctx)))
		})
}
