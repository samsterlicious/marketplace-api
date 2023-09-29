package util

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

type DefaultResponse struct {
	Message string `json:"message"`
}

func ApigatewayResponse(body string, statusCode int) (events.APIGatewayV2HTTPResponse, error) {
	return events.APIGatewayV2HTTPResponse{Body: body, Headers: map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET,POST,OPTIONS",
		"Access-Control-Allow-Headers": "X-Amz-Date,X-Api-Key,X-Amz-Security-Token,X-Requested-With,X-Auth-Token,Referer,User-Agent,Origin,Content-Type,Authorization,Accept,Access-Control-Allow-Methods,Access-Control-Allow-Origin,Access-Control-Allow-Headers",
	},
		StatusCode: statusCode}, nil
}

func GetAwsConfig(ctx context.Context) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	return cfg, err
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

func Min[T int64 | int](a, b T) T {
	if a < b {
		return a
	}
	return b
}
