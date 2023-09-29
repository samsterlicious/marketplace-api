package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"

	"sammy.link/bet"
	"sammy.link/bid"
	"sammy.link/database"
	"sammy.link/marketplace"
	"sammy.link/util"
)

type Response struct {
	util.DefaultResponse
}

func handleCreate(ctx context.Context, request events.APIGatewayV2HTTPRequest, bidService bid.Service, marketplaceService marketplace.Service) (events.APIGatewayV2HTTPResponse, error) {
	resp := Response{}

	var body = []bid.Bid{}
	betMap := make(map[string]bet.Bet)
	if err := json.Unmarshal([]byte(request.Body), &body); err != nil {
		resp.Message = err.Error()
		jsonResp, _ := json.Marshal(resp)
		fmt.Println(err.Error())
		return util.ApigatewayResponse(string(jsonResp), 500)
	}
	//replace to leagues name later
	keyName := "@TODO"
	bidService.Lock(ctx, keyName)
	user := request.RequestContext.Authorizer.JWT.Claims["https://sammy.link/email"]

	var waitGroup sync.WaitGroup

	newBidChannel := make(chan bid.Bid, 4)
	betChannel := make(chan []bet.Bet, 4)
	deleteBidChannel := make(chan []bid.Bid, 4)

	writeBids(ctx, body, user, &waitGroup, marketplaceService, bidService, betChannel, deleteBidChannel, newBidChannel, betMap)

	waitGroup.Add(1)
	handleBetsAndBidDeletes(ctx, body, betChannel, deleteBidChannel, newBidChannel, &waitGroup, bidService)
	waitGroup.Wait()

	jsonResp, _ := json.Marshal(betMap)

	bidService.ReleaseLock(ctx, keyName)

	return util.ApigatewayResponse(string(jsonResp), 200)
}

func handleBetsAndBidDeletes(ctx context.Context, bids []bid.Bid, betChannel chan []bet.Bet, deleteBidChannel chan []bid.Bid, newBidChannel chan bid.Bid, waitGroup *sync.WaitGroup, bidService bid.Service) {

	bidsAndBets := make([]bid.BidAndBet, 0)
	for _, item := range bids {
		if item.Amount > 0 && item.Amount <= 100 {
			// if time.Now().Before(item.Date) && item.Amount > 0 && item.Amount <= 100 {
			incomingBidSlice := <-deleteBidChannel
			incomingBet := <-betChannel
			newBid := <-newBidChannel
			if len(incomingBidSlice) > 0 {
				bidsAndBets = append(bidsAndBets, bid.BidAndBet{
					MyBids:   incomingBidSlice,
					IsBid:    true,
					IsDelete: true,
				})
			}

			if len(incomingBet) > 0 {
				bidsAndBets = append(bidsAndBets, bid.BidAndBet{
					MyBets: incomingBet,
					IsBid:  false,
				})
			}

			if newBid.Amount > 0 {
				bidsAndBets = append(bidsAndBets, bid.BidAndBet{
					MyBids:   []bid.Bid{newBid},
					IsBid:    true,
					IsDelete: false,
				})
			}
		}
	}

	go handleBidsAndBets(ctx, bidsAndBets, waitGroup, bidService)

}

func handleBidsAndBets(ctx context.Context, bidsAndBets []bid.BidAndBet, waitGroup *sync.WaitGroup, service bid.Service) {

	go service.WriteBidsAndBets(ctx, bidsAndBets, waitGroup)
}

func writeBids(ctx context.Context, bids []bid.Bid, user string, waitGroup *sync.WaitGroup, marketplaceService marketplace.Service, bidService bid.Service, betChannel chan []bet.Bet, deleteBidChannel chan []bid.Bid, newBidChannel chan bid.Bid, betMap map[string]bet.Bet) {

	for _, item := range bids {

		if item.Amount > 0 && item.Amount <= 100 {
			// if time.Now().Before(item.Date) && item.Amount > 0 && item.Amount <= 100 {
			item.User = user
			waitGroup.Add(1)
			go func(modifyItem bid.Bid) {
				defer waitGroup.Done()
				marketplaceService.ModifyAmount(ctx, modifyItem)
			}(item)
			item.CreateDate = time.Now()

			go func(newBid bid.Bid) {
				//query all bids for that event
				bids := bidService.GetBidsByEvent(ctx, fmt.Sprintf("%s|%s|%s|%s", newBid.Kind, newBid.Date.Format(time.RFC3339), newBid.AwayTeam, newBid.HomeTeam), newBid.Div)
				slices.SortFunc[[]bid.Bid](bids, func(bidOne bid.Bid, bidTwo bid.Bid) int {
					if bidOne.CreateDate.Before(bidTwo.CreateDate) {
						return -1
					} else if bidTwo.CreateDate.Before(bidOne.CreateDate) {
						return 1
					}
					return 0
				})
				bidsThatNeedDeleting := make([]bid.Bid, 0)

				newBets := make([]bet.Bet, 0)

				for _, existingBid := range bids {
					if existingBid.ChosenCompetitor != newBid.ChosenCompetitor && user != existingBid.User {
						lessAmount := util.Min(newBid.Amount, existingBid.Amount)

						var awayUser, homeUser string

						if newBid.ChosenCompetitor == newBid.AwayTeam {
							awayUser = user
							homeUser = existingBid.User
						} else {
							awayUser = existingBid.User
							homeUser = user
						}
						newBet := bet.Bet{
							AwayUser:         awayUser,
							HomeUser:         homeUser,
							Amount:           lessAmount,
							AwayTeam:         existingBid.AwayTeam,
							HomeTeam:         existingBid.HomeTeam,
							Status:           "PENDING",
							Spread:           existingBid.Spread,
							Kind:             existingBid.Kind,
							Date:             existingBid.Date,
							Week:             newBid.Week,
							HomeAbbreviation: newBid.HomeAbbreviation,
							AwayAbbreviation: newBid.AwayAbbreviation,
							Div:              newBid.Div,
						}
						betKey := fmt.Sprintf("%s|%s|%s|%s", awayUser, homeUser, existingBid.AwayTeam, existingBid.HomeTeam)

						if v, ok := betMap[betKey]; ok {
							v.Amount += newBet.Amount
							betMap[betKey] = v

							for i, oldBet := range newBets {
								if oldBet.AwayUser == newBet.AwayUser && oldBet.HomeUser == newBet.HomeUser && oldBet.AwayTeam == newBet.HomeTeam {
									oldBet.Amount += newBet.Amount
									newBets[i] = oldBet
								}
							}
						} else {
							betMap[betKey] = newBet
							newBets = append(newBets, newBet)
						}

						existingBid.Amount -= lessAmount
						newBid.Amount -= lessAmount

						if existingBid.Amount > lessAmount {
							existingBid.Amount -= lessAmount
							waitGroup.Add(1)
							go func(bidToUpdate bid.Bid) {
								defer waitGroup.Done()
								bidService.Update(ctx, bidToUpdate)
							}(existingBid)
						} else {
							bidsThatNeedDeleting = append(bidsThatNeedDeleting, existingBid)
						}

						if newBid.Amount == 0 {
							break
						}
					}
				}
				newBidChannel <- newBid
				betChannel <- newBets
				deleteBidChannel <- bidsThatNeedDeleting
			}(item)
		}
	}
}

func main() {
	lambda.Start(
		func(ctx context.Context, request events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
			return handleCreate(ctx, request, bid.NewService(database.GetDatabaseService[bid.DyanmoBidItem, bid.Bid](ctx)), marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
		})
}
