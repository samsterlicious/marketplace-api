package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/marketplace"
)

//NFL|2023-09-15T00:15:00Z|Vikings|Eagles

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handleCreate(ctx, events.APIGatewayV2HTTPRequest{Body: `[
		{
			"amount": 12,
			"kind": "CFB",
			"awayTeam": "Utah",
			"homeTeam": "Oregon St",
			"chosenCompetitor": "Oregon St",
			"spread": "ORST -3.0",
			"date": "2023-09-30T01:00:00Z",
			"week": 5,
			"awayAbbreviation": "UTAH",
			"homeAbbreviation": "ORST",
			"div": "default"
		}
	]`,
		RequestContext: events.APIGatewayV2HTTPRequestContext{
			Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
				JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
					// Claims: map[string]string{"https://sammy.link/email": "pgreene864@gmail.com"},
					Claims: map[string]string{"https://sammy.link/email": "sam@sam.com"},
				},
			},
		},
	},
		bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)),
		marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)),
	)
	fmt.Printf("dat resp %s", resp.Body)
}
