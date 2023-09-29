package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/marketplace"
	"sammy.link/util"
)

func update(ctx context.Context, request events.APIGatewayV2HTTPRequest, bidService bid.Service, marketplaceService marketplace.Service) (events.APIGatewayV2HTTPResponse, error) {

	var input = bid.Bid{}
	json.Unmarshal([]byte(request.Body), &input)

	fmt.Printf("%+v\n", input)
	bidService.Lock(ctx, input.Div)

	// bidService.Delete(ctx, input)
	input.Amount = -1 * input.Amount
	fmt.Println(input.Amount)
	marketplaceService.ModifyAmount(ctx, input)

	bidService.ReleaseLock(ctx, input.Div)
	jsonUser, _ := json.Marshal(map[string]string{
		"message": "success",
	})

	return util.ApigatewayResponse(string(jsonUser), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return update(ctx, request, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)),
				marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
		})
}
