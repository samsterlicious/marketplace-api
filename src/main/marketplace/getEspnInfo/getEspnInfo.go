package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/espn"
	"sammy.link/util"
)

func handle(ctx context.Context, request events.APIGatewayV2HTTPRequest, espnService espn.Service) (events.APIGatewayV2HTTPResponse, error) {

	eventId := request.QueryStringParameters["eventId"]

	espnChannel := make(chan espn.EspnResponse)

	go getEvent(eventId, "nfl", espnService, espnChannel)

	go getEvent(eventId, "college-football", espnService, espnChannel)

	resp := <-espnChannel

	espnResponse, _ := json.Marshal(resp)

	return util.ApigatewayResponse(string(espnResponse), 200)
}

func getEvent(eventId string, league string, service espn.Service, channel chan espn.EspnResponse) {
	resp, _ := service.GetEspnEvent("football", league, eventId)
	if len(resp.Sports) > 0 {
		channel <- resp
	}
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handle(ctx, request, espn.NewService(http.Client{}))
		})
}
