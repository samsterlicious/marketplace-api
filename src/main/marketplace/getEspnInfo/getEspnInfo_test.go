package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"sammy.link/espn"
)

func TestGetBetKinds(t *testing.T) {
	ctx := context.TODO()
	resp, _ := handle(ctx, events.APIGatewayV2HTTPRequest{
		QueryStringParameters: map[string]string{
			"eventId": "401520176",
		},
	}, espn.NewService(http.Client{}))

	fmt.Println(resp)
	// if len(result) != 2 || ((result[0] != "NFL" || result[1] != "CFB") && (result[1] != "NFL" || result[0] != "CFB")) {
	// 	t.Fatalf("Should return []{'NFL','CFB'} but got %s", result)
	// }
}
