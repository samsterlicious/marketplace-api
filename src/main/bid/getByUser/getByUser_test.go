package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/bid"
	"sammy.link/database"
)

func TestHandler(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handleGet(ctx, events.APIGatewayV2HTTPRequest{
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					Claims: map[string]string{"https://sammy.link/email": "pgreene864@gmail.com"},
				},
			},
		},
	}, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)))

	fmt.Println(resp.Body)
}
