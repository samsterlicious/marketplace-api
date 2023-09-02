package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"sammy.link/database"
	"sammy.link/espn"
	"sammy.link/marketplace"
	"sammy.link/util"
)

func main() {
	lambda.Start(func(ctx context.Context) {
		handler(ctx, marketplace.NewService(database.GetDatabaseService[marketplace.MarketplaceDynamoDbItem, marketplace.MarketplaceItem](ctx)))
	})
}

func handler(ctx context.Context, service marketplace.Service) {

	marketplaceDbItems := service.GetItems(ctx)

	fmt.Printf("%d is len", len(marketplaceDbItems))

	marketplaceCache := make(map[string]bool)

	for _, item := range marketplaceDbItems {
		marketplaceCache[marketplace.BuildMarketplaceDynamoId(item.Kind, item.Date, item.AwayTeam, item.HomeTeam)] = true
	}

	espnResponseChannel := make(chan espn.EspnResponse, 4)

	sportMap := map[string][]string{
		"football": {"nfl", "college-football"},
	}

	firstDate := time.Now()
	secondDate := firstDate.AddDate(0, 0, 7)
	//@TODO use the espn service from the handler parameter
	espnService := espn.NewService(http.Client{})
	for sport, leagues := range sportMap {
		for _, league := range leagues {
			go espnService.GetEspnData(sport, league, espnResponseChannel, firstDate, secondDate)
		}
	}

	events := make([]marketplace.MarketplaceItem, 0, 25)

	eventBridgeMap := make(map[int64][]marketplace.MarketplaceItem)

	for _, leagues := range sportMap {
		for range leagues {
			response := <-espnResponseChannel

			for _, sport := range response.Sports {
				for _, league := range sport.Leagues {
					for _, event := range league.Events {

						awayTeam := event.Competitors[0]
						homeTeam := event.Competitors[1]

						kind := getKind(league.Name)

						if !marketplaceCache[marketplace.BuildMarketplaceDynamoId(kind, event.Date, awayTeam.Name, homeTeam.Name)] && event.Date.After(time.Now()) && event.Odds.Details != "" && event.Odds.Details != "OFF" {
							item := marketplace.MarketplaceItem{AwayTeam: awayTeam.Name,
								HomeTeam: homeTeam.Name, Date: event.Date, Kind: kind,
								Spread: event.Odds.Details}

							formattedDate := event.Date.Unix()
							if val, ok := eventBridgeMap[formattedDate]; ok {
								eventBridgeMap[formattedDate] = append(val, item)
							} else {
								eventBridgeMap[formattedDate] = []marketplace.MarketplaceItem{item}
							}

							events = append(events, item)
						}
					}
				}
			}
		}
	}

	fmt.Printf("saving %d events\n", len(events))
	var waitGroup sync.WaitGroup

	for i := 0; i < len(events); i += 25 {
		waitGroup.Add(1)
		go func(myEvents []marketplace.MarketplaceItem) {
			defer waitGroup.Done()
			service.Write(ctx, myEvents)
		}(events[i:util.Min(i+25, len(events))])
	}

	config, _ := util.GetAwsConfig(ctx)
	bridge := eventbridge.NewFromConfig(config)

	nowUnix := time.Now().Unix()

	for eventDate, events := range eventBridgeMap {

		waitGroup.Add(1)

		go func(myEventDate int64, myEvents []marketplace.MarketplaceItem, myNowUnix int64) {
			defer waitGroup.Done()
			createRuleAndTarget(ctx, bridge, myEventDate, nowUnix, myEvents)

		}(eventDate, events, nowUnix)

	}

	waitGroup.Wait()
}

func createRuleAndTarget(ctx context.Context, bridge *eventbridge.Client, myEventDate int64, myNowUnix int64, myEvents []marketplace.MarketplaceItem) {
	ruleName := aws.String(fmt.Sprintf("%d-%d", myEventDate, myNowUnix))

	firstEventDate := myEvents[0].Date
	cronExpression := fmt.Sprintf("cron(%d %d %d %d ? %d)", firstEventDate.Minute(), firstEventDate.Hour(), firstEventDate.Day(), int(firstEventDate.Month()), firstEventDate.Year())

	input := &eventbridge.PutRuleInput{
		Name:               ruleName,
		ScheduleExpression: aws.String(cronExpression),
		State:              types.RuleStateEnabled,
	}

	_, err := bridge.PutRule(ctx, input)

	if err != nil {
		fmt.Println(err.Error())
	}

	jsonInput, _ := json.Marshal(TargetInput{
		RuleName: *ruleName,
		Events:   myEvents,
	})

	targetInput := &eventbridge.PutTargetsInput{
		Rule: ruleName,
		Targets: []types.Target{
			{
				Arn:   aws.String(os.Getenv("CREATE_LAMBDA_ARN")),
				Id:    aws.String("my-target"),
				Input: aws.String(string(jsonInput)),
			},
		},
	}

	bridge.PutTargets(ctx, targetInput)
}

func getKind(league string) string {
	if league == "NCAA - Football" {
		return "CFB"
	}
	return "NFL"
}

type TargetInput struct {
	Events   []marketplace.MarketplaceItem `json:"events"`
	RuleName string                        `json:"ruleName"`
}
