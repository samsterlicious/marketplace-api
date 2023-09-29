package main

/*
{
    "div": "default",
    "awayUser": "pgreene864@gmail.com",
    "homeUser": "",
    "amount": "15",
    "awayTeam": "Bills",
    "homeTeam": "Jets",
    "status": "BAD",
    "spread": "BUF -2.5",
    "kind": "NFL",
    "week": 0,
    "createDate": "2023-09-11T17:58:55Z",
    "date": "2023-09-12T00:15:00Z",
    "homeAbbreviation": "NYJ",
    "awayAbbreviation": "BUF",
    "chosenCompetitor": "Bills",
    "user": "pgreene864@gmail.com"
}
*/

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/marketplace"
)

func TestUpdate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := update(ctx, events.APIGatewayV2HTTPRequest{
		Body: `{
			"div": "default",
			"awayUser": "",
			"homeUser": "pgreene864@gmail.com",
			"amount": 0,
			"awayTeam": "Utah",
			"homeTeam": "Oregon St",
			"status": "BAD",
			"spread": "ORST -3.0",
			"kind": "CFB",
			"week": 5,
			"createDate": "2023-09-29T15:38:39Z",
			"date": "2023-09-30T01:00:00Z",
			"homeAbbreviation": "ORST",
			"awayAbbreviation": "UTAH",
			"chosenCompetitor": "Oregon St",
			"user": "pgreene864@gmail.com"
		}`,
	}, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)),
		marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
	fmt.Printf("your boy %s", resp.Body)
}
