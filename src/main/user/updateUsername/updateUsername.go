package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/database"
	"sammy.link/league"
	"sammy.link/user"
	"sammy.link/util"
)

type Input struct {
	Name string `json:"name"`
	Div  string `json:"div"`
}

func update(ctx context.Context, request events.APIGatewayV2HTTPRequest, userService user.Service, leagueService league.Service) (events.APIGatewayV2HTTPResponse, error) {

	var input = Input{}
	json.Unmarshal([]byte(request.Body), &input)
	email := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]

	users := leagueService.GetUsers(ctx, input.Div)

	fmt.Printf("sam %+v", input)
	for _, existingUser := range users {
		if existingUser.Name == input.Name {
			return util.ApigatewayResponse("", 500)
		}
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	go func() {
		defer waitGroup.Done()
		userService.UpdateName(ctx,
			user.Item{
				Email:  email,
				Name:   input.Name,
				League: input.Div,
			},
		)
	}()

	go func() {
		defer waitGroup.Done()
		leagueService.UpdateUserName(ctx, input.Div, email, input.Name)
	}()

	waitGroup.Wait()

	jsonUser, _ := json.Marshal(map[string]string{
		"message": "success",
	})

	return util.ApigatewayResponse(string(jsonUser), 200)
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return update(ctx, request, user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)),
				league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
					database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
		})
}
