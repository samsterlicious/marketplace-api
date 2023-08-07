package util

import (
	"github.com/aws/aws-lambda-go/events"
)

func ApigatewayResponse(body string) (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{Body: body, Headers: map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "X-Amz-Date,X-Api-Key,X-Amz-Security-Token,X-Requested-With,X-Auth-Token,Referer,User-Agent,Origin,Content-Type,Authorization,Accept,Access-Control-Allow-Methods,Access-Control-Allow-Origin,Access-Control-Allow-Headers",
	},
		StatusCode: 200}, nil
}

func Filter[T any](slice []T, test func(T) bool) []T {
	ret := make([]T, 0, len(slice))
	for _, element := range slice {
		if test(element) {
			ret = append(ret, element)
		}
	}
	return ret
}
