package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/marketplace"
	"sammy.link/util"
)

type Response struct {
	util.DefaultResponse
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func handleCreate(ctx context.Context, request events.APIGatewayV2HTTPRequest, bidService bid.Service, marketplaceService marketplace.Service) (events.APIGatewayV2HTTPResponse, error) {
	resp := Response{}

	var body = []bid.Bid{}

	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		resp.Message = err.Error()
		jsonResp, _ := json.Marshal(resp)

		return util.ApigatewayResponse(string(jsonResp), 500)
	}

	user := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]

	var waitGroup sync.WaitGroup

	goodBids := make([]bid.Bid, 0, 25)

	for _, item := range body {

		if time.Now().Before(item.Date) && item.Amount > 0 {
			item.User = user
			waitGroup.Add(1)
			go func(modifyItem bid.Bid) {
				defer waitGroup.Done()
				marketplaceService.ModifyAmount(ctx, modifyItem)
			}(item)
			item.CreateDate = time.Now()
			goodBids = append(goodBids, item)
		}
	}

	for i := 0; i < len(goodBids); i += 25 {
		waitGroup.Add(1)
		go func(bidSlice []bid.Bid) {
			defer waitGroup.Done()
			bidService.WriteBids(ctx, bidSlice)
		}(goodBids[i:min(i+25, len(goodBids))])
	}

	waitGroup.Wait()

	resp.Message = "success"

	jsonResp, _ := json.Marshal(resp)

	return util.ApigatewayResponse(string(jsonResp), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleCreate(ctx, request, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)), marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
		})
}
