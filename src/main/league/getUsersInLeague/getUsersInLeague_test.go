package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/database"
	"sammy.link/league"
)

func TestUpdate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handleGet(ctx, events.APIGatewayV2HTTPRequest{
		PathParameters: map[string]string{
			"league": "default",
		},
	}, league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
		database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
	fmt.Printf("your boy %s", resp.Body)
}
