package main

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"sammy.link/bid"
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

var dynamoClient *dynamodb.Client

func handleCreate(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	resp := Response{}

	config := util.GetAwsConfig(ctx)
	if dynamoClient == nil {
		dynamoClient = util.GetDynamoClient(config)
	}

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
				marketplace.ModifyAmount(ctx, modifyItem, dynamoClient)
			}(item)
			item.CreateDate = time.Now()
			goodBids = append(goodBids, item)
		}
	}

	for i := 0; i < len(goodBids); i += 25 {
		waitGroup.Add(1)
		go func(bidSlice []bid.Bid) {
			defer waitGroup.Done()
			bid.WriteBids(ctx, bidSlice, dynamoClient)
		}(goodBids[i:min(i+25, len(goodBids))])
	}

	waitGroup.Wait()

	resp.Message = "success"

	jsonResp, _ := json.Marshal(resp)

	return util.ApigatewayResponse(string(jsonResp), 200)
}

func main() {
	lambda.Start(handleCreate)
}
