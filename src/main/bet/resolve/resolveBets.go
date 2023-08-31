package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"sammy.link/bet"
	"sammy.link/espn"
	"sammy.link/outcome"
	"sammy.link/util"
)

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) {
	fivehours, _ := time.ParseDuration("-5h")
	dynamoClient := util.GetDynamoClient(util.GetAwsConfig(ctx))

	yesterday := time.Now().Add(fivehours)
	dynamoBets := bet.GetBetsByEventDate(ctx, yesterday.Format("20060102"), dynamoClient)

	if len(dynamoBets) > 0 {
		bets := bet.ConvertDyanmoList(dynamoBets)
		kinds := getBetKinds(bets)
		spreads := getSpreads(bets)

		outcomeChan := make(chan []outcome.Outcome, len(kinds))

		for _, kind := range kinds {

			go func(myKind string, myBets []bet.Bet, myOutcomeChan chan []outcome.Outcome) {
				kindResp := espn.ConvertKind(myKind)
				espnResponseChannel := make(chan espn.EspnResponse)
				espn.GetEspnData(kindResp.Sport, kindResp.League, espnResponseChannel, yesterday, yesterday)
				espnResp := <-espnResponseChannel
				winners := getWinners(spreads, espnResp)
				outcomeSlice := make([]outcome.Outcome, 0, len(myBets))

				for _, bet := range bets {
					if bet.Kind == myKind {
						gameName := fmt.Sprintf("%s|%s", bet.AwayTeam, bet.HomeTeam)
						gameResult := winners[gameName]

						if gameResult.team == bet.AwayTeam {
							outcomeSlice = append(outcomeSlice, outcome.Outcome{
								Winner:  bet.AwayUser,
								Loser:   bet.HomeUser,
								EventId: gameResult.eventId,
								Week:    gameResult.week,
								Amount:  bet.Amount,
							})
						} else if gameResult.team == bet.HomeTeam {
							outcomeSlice = append(outcomeSlice, outcome.Outcome{
								Winner:  bet.HomeUser,
								Loser:   bet.AwayUser,
								EventId: gameResult.eventId,
								Week:    gameResult.week,
								Amount:  bet.Amount,
							})
						}

					}
				}
				myOutcomeChan <- outcomeSlice
			}(kind, bets, outcomeChan)
		}

		var waitGroup sync.WaitGroup

		for range kinds {
			outcomeSlice := <-outcomeChan
			for i := 0; i < len(outcomeSlice); i += 25 {
				waitGroup.Add(1)
				go func(mySlice []outcome.Outcome) {
					defer waitGroup.Done()
					outcome.Write(ctx, mySlice, dynamoClient)
				}(outcomeSlice[i:util.Min(i+25, len(outcomeSlice))])
			}
		}

		waitGroup.Wait()

	}
}

func getBetKinds(bets []bet.Bet) []string {
	kindMap := make(map[string]bool)

	for _, bet := range bets {
		kindMap[bet.Kind] = true
	}

	stringSlice := make([]string, 2)

	i := 0
	for key := range kindMap {
		stringSlice[i] = key
		i++
	}

	return stringSlice
}

func getSpreads(bets []bet.Bet) map[string]string {
	resp := make(map[string]string)

	for _, bet := range bets {
		resp[fmt.Sprintf("%s|%s", bet.AwayTeam, bet.HomeTeam)] = bet.Spread
	}

	return resp
}

func getWinners(spreads map[string]string, espnResp espn.EspnResponse) map[string]winner {

	winnersMap := make(map[string]winner)

	for _, sport := range espnResp.Sports {
		for _, league := range sport.Leagues {
			for _, event := range league.Events {
				awayTeam := event.Competitors[0]
				homeTeam := event.Competitors[1]

				awayScore, _ := strconv.ParseFloat(awayTeam.Score, 64)
				homeScore, _ := strconv.ParseFloat(homeTeam.Score, 64)
				gameName := fmt.Sprintf("%s|%s", awayTeam.Name, homeTeam.Name)
				spread := spreads[gameName]

				spreadRegexp := regexp.MustCompile(`^(?P<Team>\S+)\s+(?P<Points>\S+)$`)

				doesMatch := spreadRegexp.MatchString(spread)

				if doesMatch {
					paramsMap := make(map[string]string)
					match := spreadRegexp.FindStringSubmatch(spread)
					for i, name := range spreadRegexp.SubexpNames() {
						if i > 0 && i <= len(match) {
							paramsMap[name] = match[i]
						}
					}

					points, _ := strconv.ParseFloat(paramsMap["Points"], 64)
					if paramsMap["Team"] == awayTeam.Abbreviation {
						awayScore += points
					} else {
						homeScore += points
					}
				}
				if awayScore > homeScore {
					winnersMap[gameName] = winner{
						team:    awayTeam.Name,
						eventId: event.Id,
						week:    fmt.Sprintf("%d", event.Week),
					}
				} else if homeScore > awayScore {
					winnersMap[gameName] = winner{
						team:    homeTeam.Name,
						eventId: event.Id,
						week:    fmt.Sprintf("%d", event.Week),
					}
				} else {
					winnersMap[gameName] = winner{
						team:    "",
						eventId: "",
						week:    fmt.Sprintf("%d", event.Week),
					}
				}
			}
		}
	}
	return winnersMap
}

type winner struct {
	team    string
	eventId string
	week    string
}
