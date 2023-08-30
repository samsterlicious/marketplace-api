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
	"sammy.link/marketplace"
	"sammy.link/util"
)

type espnResponse struct {
	Sports []espnSport `json:"sports"`
}

type espnSport struct {
	Name    string       `json:"name"`
	Leagues []espnLeague `json:"leagues"`
}

type espnLeague struct {
	Name         string      `json:"name"`
	Abbreviation string      `json:"abbreviation"`
	Events       []espnEvent `json:"events"`
}

type espnEvent struct {
	Name        string           `json:"name"`
	ShortName   string           `json:"shortName"`
	Date        time.Time        `json:"date"`
	Odds        espnOdds         `json:"odds"`
	Status      string           `json:"status"`
	Competitors []espnCompetitor `json:"competitors"`
}

type espnCompetitor struct {
	Name     string `json:"name"`
	HomeAway string `json:"homeAway"`
	Winner   bool   `json:"winner"`
}

type espnOdds struct {
	Details string `json:"details"`
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func getEspnData(sport string, league string, channel chan espnResponse) {
	const format = "20060102"

	resp, err := http.Get(fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=%s&league=%s&dates=%s-%s",
		sport, league, time.Now().Format(format), time.Now().AddDate(0, 0, 7).Format(format)))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var response espnResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		panic(err)
	}

	channel <- response
}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) {

	config := util.GetAwsConfig(ctx)
	client := util.GetDynamoClient(config)
	marketplaceDbItems := marketplace.GetMarketplaceItems(ctx, client)

	fmt.Printf("%d is len", len(marketplaceDbItems))

	marketplaceCache := make(map[string]bool)

	for _, item := range marketplaceDbItems {
		marketplaceCache[item.SortKey] = true
	}

	espnResponseChannel := make(chan espnResponse, 4)

	sportMap := map[string][]string{
		"football": {"nfl", "college-football"},
	}

	for sport, leagues := range sportMap {
		for _, league := range leagues {
			go getEspnData(sport, league, espnResponseChannel)
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
			marketplace.WriteMarketplaceItems(ctx, myEvents, &waitGroup, client)
		}(events[i:min(i+25, len(events))])
	}

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
