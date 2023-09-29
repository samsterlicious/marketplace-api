package main

import (
	"context"
	"encoding/json"
	"slices"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bet"
	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/util"
)

func handleGet(ctx context.Context, request events.APIGatewayV2HTTPRequest, betService bet.Service, bidService bid.Service) (events.APIGatewayV2HTTPResponse, error) {

	var bets []bet.Bet
	div := request.QueryStringParameters["div"]

	if week, ok := request.PathParameters["date"]; ok {
		bets = betService.GetBetsByWeek(ctx, div, week)
	} else {
		user := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]
		betChannel := make(chan []bet.Bet, 3)

		go func() {
			myBets := betService.GetBetsByUser(ctx, user, true)
			betChannel <- myBets
		}()

		go func() {
			myBets := betService.GetBetsByUser(ctx, user, false)
			betChannel <- myBets
		}()

		go func() {
			bids := bidService.GetBidsByUser(ctx, user)

			notBets := make([]bet.Bet, len(bids))

			for i, badBid := range bids {
				notBets[i] = bet.Bet{
					Div:              badBid.Div,
					Amount:           badBid.Amount,
					Status:           "BAD",
					AwayTeam:         badBid.AwayTeam,
					HomeTeam:         badBid.HomeTeam,
					Spread:           badBid.Spread,
					Kind:             badBid.Kind,
					Week:             badBid.Week,
					CreateDate:       badBid.CreateDate,
					Date:             badBid.Date,
					HomeAbbreviation: badBid.HomeAbbreviation,
					AwayAbbreviation: badBid.AwayAbbreviation,
				}

				if badBid.ChosenCompetitor == badBid.AwayTeam {
					notBets[i].AwayUser = user
				} else {
					notBets[i].HomeUser = user
				}
			}
			betChannel <- notBets
		}()

		for i := 0; i < 3; i++ {
			channelBets := <-betChannel
			bets = append(bets, channelBets...)
		}

	}
	slices.SortFunc[[]bet.Bet](bets, func(betOne bet.Bet, betTwo bet.Bet) int {
		if betTwo.Date.Before(betOne.Date) {
			return -1
		} else if betOne.Date.Before(betTwo.Date) {
			return 1
		}
		return 0
	})
	jsonBets, _ := json.Marshal(bets)
	return util.ApigatewayResponse(string(jsonBets), 200)
}

func main() {
	lambda.Start(func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
		return handleGet(ctx, request, bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)),
			bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)))
	})
}
