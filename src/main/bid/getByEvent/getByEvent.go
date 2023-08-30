package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"sammy.link/bid"
	"sammy.link/util"
)

var dynamoClient *dynamodb.Client

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	if dynamoClient == nil {
		dynamoClient = util.GetDynamoClient(util.GetAwsConfig(ctx))
	}

	fmt.Println(request.PathParameters["event"])
	dynamoBids := bid.GetBidsByEvent(ctx, request.PathParameters["event"], dynamoClient)

	bids := make([]bid.Bid, 0, len(dynamoBids))

	for _, bid := range dynamoBids {
		bids = append(bids, bid.GetItem())
	}

	jsonBids, _ := json.Marshal(bids)

	return util.ApigatewayResponse(string(jsonBids), 200)
}

func main() {
	lambda.Start(handleGet)
}
