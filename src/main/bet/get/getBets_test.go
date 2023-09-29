package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/bet"
	"sammy.link/bid"
	"sammy.link/database"
)

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handleGet(ctx, events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{
			"date": "1",
		},
		QueryStringParameters: map[string]string{
			"div": "default",
		},
	}, bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)),
		bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)))
	fmt.Printf("your boy %s", resp.Body)
}
