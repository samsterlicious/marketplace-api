package main

import (
	"context"
	"testing"

	"sammy.link/database"
	"sammy.link/marketplace"
)

func TestCreate(t *testing.T) {
	ctx := context.TODO()
	handler(ctx, marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
}
