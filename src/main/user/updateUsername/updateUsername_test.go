package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/database"
	"sammy.link/league"
	"sammy.link/user"
)

func TestUpdate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := update(ctx, events.APIGatewayV2HTTPRequest{
		Body: "{\"name\": \"samg\", \"div\": \"default\"}",
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
				Claims: map[string]string{
					"https://sammy.link/email": "pgreene864@gmail.com.com",
				},
			}}},
	}, user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)),
		league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
			database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
	fmt.Printf("your boy %s", resp.Body)
}
