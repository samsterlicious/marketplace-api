package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/database"
	"sammy.link/marketplace"
)

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handleGet(ctx, events.APIGatewayV2HTTPRequest{}, marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
	fmt.Printf("your boy %s", resp.Body)
}
