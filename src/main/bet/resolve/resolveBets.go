package main

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid"
	"sammy.link/bet"
	"sammy.link/database"
	"sammy.link/espn"
	"sammy.link/outcome"
	"sammy.link/user"
	"sammy.link/util"
)

func main() {
	lambda.Start(
		func(ctx context.Context) {
			handler(ctx, outcome.NewService(database.GetDatabaseService[outcome.OutcomeDynamoItem, outcome.OutcomeItem](ctx)),
				bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)), espn.NewService(http.Client{}),
				user.NewService(database.GetDatabaseService[user.DynamoItem, user.Item](ctx)))
		})
}

func handler(ctx context.Context, outcomeService outcome.Service, betService bet.Service, espnServce espn.Service, userService user.Service) {
	fivehours, _ := time.ParseDuration("-5h")
	yesterday := time.Now().Add(fivehours)
	bets := betService.GetBetsByEventDate(ctx, yesterday.Format("20060102"))

	if len(bets) > 0 {
		kinds := getBetKinds(bets)
		spreads := getSpreads(bets)

		outcomeChan := make(chan []outcome.OutcomeItem, len(kinds))
		userChannel := make(chan user.Item, 10)

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

						if gameResult.team == bet.AwayTeam {
							outcomeSlice = append(outcomeSlice, outcome.OutcomeItem{
								Winner:  bet.AwayUser,
								Loser:   bet.HomeUser,
								EventId: gameResult.eventId,
								Week:    gameResult.week,
								Amount:  bet.Amount,
								Id:      uuid.NewString(),
							})
							sendUpdateUserInfo(bet.AwayUser, bet.HomeUser, bet.Amount, userChannel)
						} else if gameResult.team == bet.HomeTeam {
							outcomeSlice = append(outcomeSlice, outcome.OutcomeItem{
								Winner:  bet.HomeUser,
								Loser:   bet.AwayUser,
								EventId: gameResult.eventId,
								Week:    gameResult.week,
								Amount:  bet.Amount,
								Id:      uuid.NewString(),
							})
							sendUpdateUserInfo(bet.HomeUser, bet.AwayUser, bet.Amount, userChannel)
						} else {
							sendUpdateUserInfo("", "", 0, userChannel)
						}

					}

				}
				myOutcomeChan <- outcomeSlice
			}(kind, bets, outcomeChan)
		}

		var waitGroup sync.WaitGroup

		writeOutcomes(ctx, kinds, outcomeChan, &waitGroup, outcomeService)

		waitGroup.Add(1)
		go updateUserTotals(ctx, userService, userChannel, bets, &waitGroup)
		waitGroup.Wait()

	}

}

func sendUpdateUserInfo(winner string, loser string, amount int64, userChannel chan user.Item) {
	userChannel <- user.Item{
		Email: winner,
		Total: amount,
	}
	userChannel <- user.Item{
		Email: loser,
		Total: -1 * amount,
	}
}

func updateUserTotals(ctx context.Context, userService user.Service, userChannel chan user.Item, bets []bet.Bet, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()
	userMap := make(map[string]int64)
	for i := 0; i < len(bets)*2; i++ {
		updateUser := <-userChannel
		if updateUser.Email == "" {
			continue
		}
		if existingUser, ok := userMap[updateUser.Email]; ok {
			userMap[updateUser.Email] = existingUser + updateUser.Total
		} else {
			userMap[updateUser.Email] = updateUser.Total
		}
	}

	for userEmail, total := range userMap {
		if total != 0 {
			fmt.Printf("email %s; total %d\n", userEmail, total)
			waitGroup.Add(1)
			go func(email string, amount int64) {
				defer waitGroup.Done()
				userService.Update(ctx, email, amount)
			}(userEmail, total)
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
