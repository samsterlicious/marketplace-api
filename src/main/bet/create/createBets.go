package main

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"sammy.link/bet"
	"sammy.link/bid"
	"sammy.link/marketplace"
	"sammy.link/util"
)

var dynamoClient *dynamodb.Client

func handler(ctx context.Context, message json.RawMessage) error {
	input := TargetInput{}

	fmt.Println(message)

	err := json.Unmarshal(message, &input)

	if err != nil {
		fmt.Println(err.Error())
	}

	fmt.Printf("my name is %s and the len is %d\n", input.RuleName, len(input.Events))

	var waitGroup sync.WaitGroup

	//loops through events

	config := util.GetAwsConfig(ctx)

	if dynamoClient == nil {
		dynamoClient = util.GetDynamoClient(config)
	}

	betSliceChannel := make(chan []bet.Bet)

	for _, event := range input.Events {
		waitGroup.Add(1)
		go func(myEvent marketplace.MarketplaceItem) {
			defer waitGroup.Done()
			createBets(ctx, myEvent, dynamoClient, betSliceChannel)
		}(event)

	}

	bridge := eventbridge.NewFromConfig(config)

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()

		ruleName := aws.String(input.RuleName)

		_, err := bridge.RemoveTargets(ctx, &eventbridge.RemoveTargetsInput{
			Ids:  []string{"my-target"},
			Rule: ruleName,
		})

		if err != nil {
			fmt.Println(err.Error())
		}

		_, err = bridge.DeleteRule(ctx, &eventbridge.DeleteRuleInput{
			Name: ruleName,
		})

		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	bigDaddyBetSlice := make([]bet.Bet, 0, 25)

	for range input.Events {
		betSlice := <-betSliceChannel
		bigDaddyBetSlice = append(bigDaddyBetSlice, betSlice...)
	}
	now := time.Now()
	for i := 0; i < len(bigDaddyBetSlice); i += 25 {
		waitGroup.Add(1)
		go func(myBets []bet.Bet) {
			defer waitGroup.Done()
			bet.WriteBets(ctx, myBets, dynamoClient, now)
		}(bigDaddyBetSlice[i:min(int64(i+25), int64(len(bigDaddyBetSlice)))])
	}
	waitGroup.Wait()
	return nil
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func createBets(ctx context.Context, event marketplace.MarketplaceItem, client *dynamodb.Client, betSliceChannel chan []bet.Bet) {
	dynamoBids := bid.GetBidsByEvent(ctx, fmt.Sprintf("%s|%s|%s|%s", event.Kind, event.Date.Format(time.RFC3339), event.AwayTeam, event.HomeTeam), client)

	bidMap := make(map[string][]bid.Bid)

	var awayTeam string
	var homeTeam string
	var kind string
	var date time.Time

	for _, dynamoBid := range dynamoBids {
		bidItem := dynamoBid.GetItem()
		if k, ok := bidMap[bidItem.ChosenCompetitor]; ok {
			bidMap[bidItem.ChosenCompetitor] = append(k, bidItem)
		} else {
			awayTeam = bidItem.AwayTeam
			homeTeam = bidItem.HomeTeam
			kind = bidItem.Kind
			date = bidItem.Date
			bidMap[bidItem.ChosenCompetitor] = []bid.Bid{bidItem}
		}
	}

	sliceLength := 1

	for _, bids := range bidMap {
		sliceLength = len(bids)
		slices.SortFunc[[]bid.Bid](bids, func(bidOne bid.Bid, bidTwo bid.Bid) int {
			if bidOne.CreateDate.Before(bidTwo.CreateDate) {
				return -1
			} else if bidTwo.CreateDate.Before(bidOne.CreateDate) {
				return 1
			}
			return 0
		})
	}

	bets := make([]bet.Bet, 0, sliceLength)

	homeIndex := 0

	awaySlice := bidMap[awayTeam]
	homeSlice := bidMap[homeTeam]

	usersBetMap := make(map[string]int)

	for awayIndex := 0; awayIndex < len(awaySlice); {
		if homeIndex < len(homeSlice) {

			usersBetKey := fmt.Sprintf("%s|%s", awaySlice[awayIndex].User, homeSlice[homeIndex].User)

			lessAmount := min(awaySlice[awayIndex].Amount, homeSlice[homeIndex].Amount)

			if existingBetIndex, ok := usersBetMap[usersBetKey]; ok {
				bets[existingBetIndex].Amount += lessAmount
			} else {
				fmt.Printf("else\n")
				newBet := bet.Bet{
					AwayUser: awaySlice[awayIndex].User,
					HomeUser: homeSlice[homeIndex].User,
					Amount:   lessAmount,
					AwayTeam: awayTeam,
					HomeTeam: homeTeam,
					Status:   "PENDING",
					Spread:   awaySlice[awayIndex].Spread,
					Kind:     kind,
					Date:     date,
				}
				bets = append(bets, newBet)

				usersBetMap[usersBetKey] = len(bets) - 1
			}

			awaySlice[awayIndex].Amount -= lessAmount
			homeSlice[homeIndex].Amount -= lessAmount

			if homeSlice[homeIndex].Amount == 0 {
				homeIndex++
			}

			if awaySlice[awayIndex].Amount == 0 {
				awayIndex++
			}
		} else {
			break
		}
	}

	betSliceChannel <- bets
}

func main() {
	lambda.Start(handler)
}

type TargetInput struct {
	Events   []marketplace.MarketplaceItem `json:"events"`
	RuleName string                        `json:"ruleName"`
}
