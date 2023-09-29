package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"sammy.link/bet"
	"sammy.link/database"
	"sammy.link/espn"
	"sammy.link/league"
	"sammy.link/outcome"
	"sammy.link/util"
)

func main() {
	lambda.Start(
		func(ctx context.Context) {
			handler(ctx, outcome.NewService(database.GetDatabaseService[outcome.OutcomeDynamoItem, outcome.OutcomeItem](ctx)),
				bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)), espn.NewService(http.Client{}),
				league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
					database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
		})
}

func handler(ctx context.Context, outcomeService outcome.Service, betService bet.Service, espnServce espn.Service, leagueService league.Service) {
	fivehours, _ := time.ParseDuration("-5h")
	yesterday := time.Now().Add(fivehours)
	bets := betService.GetBetsByEventDate(ctx, yesterday.Format("20060102"))
	twentyFourHours, _ := time.ParseDuration("-24h")
	yesterday = time.Now().Add(twentyFourHours)
	if len(bets) > 0 {
		kinds := getBetKinds(bets)
		spreads := getSpreads(bets)

		outcomeChan := make(chan []outcome.OutcomeItem, len(kinds))
		userLeagueChannel := make(chan league.UserInLeagueItem, 10)

		for _, kind := range kinds {

			go func(myKind string, myBets []bet.Bet, myOutcomeChan chan []outcome.OutcomeItem) {
				kindResp := espnServce.ConvertKind(myKind)
				espnResponseChannel := make(chan espn.EspnResponse, 1)
				espnServce.GetEspnData(kindResp.Sport, kindResp.League, espnResponseChannel, yesterday, yesterday)
				espnResp := <-espnResponseChannel
				winners := getWinners(spreads, espnResp)
				outcomeSlice := make([]outcome.OutcomeItem, 0, len(myBets))

				for _, bet := range bets {
					if bet.Kind == myKind {
						gameName := fmt.Sprintf("%s|%s", bet.AwayTeam, bet.HomeTeam)
						gameResult := winners[gameName]
						week, _ := strconv.Atoi(gameResult.week)
						if gameResult.team == bet.AwayTeam {
							outcomeSlice = append(outcomeSlice, outcome.OutcomeItem{
								Winner:  bet.AwayUser,
								Loser:   bet.HomeUser,
								EventId: gameResult.eventId,
								Week:    week,
								Amount:  bet.Amount,
								Id:      uuid.NewString(),
								Div:     bet.Div,
							})
							sendUpdateUserInfo(bet.AwayUser, bet.HomeUser, bet.Amount, bet.Div, userLeagueChannel)
						} else if gameResult.team == bet.HomeTeam {
							outcomeSlice = append(outcomeSlice, outcome.OutcomeItem{
								Winner:  bet.HomeUser,
								Loser:   bet.AwayUser,
								EventId: gameResult.eventId,
								Week:    week,
								Amount:  bet.Amount,
								Id:      uuid.NewString(),
								Div:     bet.Div,
							})
							sendUpdateUserInfo(bet.HomeUser, bet.AwayUser, bet.Amount, bet.Div, userLeagueChannel)
						} else {
							sendUpdateUserInfo("", "", 0, bet.Div, userLeagueChannel)
						}

					}

				}
				myOutcomeChan <- outcomeSlice
			}(kind, bets, outcomeChan)
		}

		var waitGroup sync.WaitGroup

		writeOutcomes(ctx, kinds, outcomeChan, &waitGroup, outcomeService)

		waitGroup.Add(1)
		go updateUserTotals(ctx, leagueService, userLeagueChannel, bets, &waitGroup)
		waitGroup.Wait()

	}

}

func sendUpdateUserInfo(winner string, loser string, amount int64, div string, userLeagueChannel chan league.UserInLeagueItem) {
	userLeagueChannel <- league.UserInLeagueItem{
		Email:  winner,
		Total:  amount,
		League: div,
	}
	userLeagueChannel <- league.UserInLeagueItem{
		Email:  loser,
		Total:  -1 * amount,
		League: div,
	}
}

func updateUserTotals(ctx context.Context, leagueService league.Service, userLeagueChannel chan league.UserInLeagueItem, bets []bet.Bet, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	userMap := make(map[string]int64)
	for i := 0; i < len(bets)*2; i++ {
		updateUser := <-userLeagueChannel
		if updateUser.Email == "" {
			continue
		}
		key := fmt.Sprintf("%s|%s", updateUser.League, updateUser.Email)
		if existingUser, ok := userMap[key]; ok {
			userMap[key] = existingUser + updateUser.Total
		} else {
			userMap[key] = updateUser.Total
		}
	}

	for userLeagueAndEmail, total := range userMap {
		if total != 0 {
			waitGroup.Add(1)
			splits := strings.Split(userLeagueAndEmail, "|")
			go func(league string, email string, amount int64) {
				defer waitGroup.Done()
				leagueService.UpdateUserAmount(ctx, league, email, amount)
			}(splits[0], splits[1], total)
		}
	}
}

func writeOutcomes(ctx context.Context, kinds []string, outcomeChan chan []outcome.OutcomeItem, waitGroup *sync.WaitGroup, o outcome.Service) {
	for range kinds {
		outcomeSlice := <-outcomeChan
		for i := 0; i < len(outcomeSlice) && len(outcomeSlice) > 0; i += 25 {
			waitGroup.Add(1)
			go func(mySlice []outcome.OutcomeItem) {
				defer waitGroup.Done()
				o.Write(ctx, mySlice)
			}(outcomeSlice[i:util.Min(i+25, len(outcomeSlice))])
		}
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
