package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"samster.link/marketplace"

	"github.com/aws/aws-lambda-go/lambda"
)

type espnResponse struct {
	Sports []espnSport `json:"sports"`
}

type espnSport struct {
	Name    string       `json:"name"`
	Logos   []espnLogo   `json:"logos"`
	Leagues []espnLeague `json:"leagues"`
}

type espnLogo struct {
	Href string `json:"href"`
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
	Logo     string `json:"logo"`
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

func getEspnData(sport string, channel chan espnResponse) {
	format := "20060102"

	resp, err := http.Get(fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=%s&dates=%s-%s",
		sport, time.Now().Format(format), time.Now().AddDate(0, 0, 7).Format(format)))

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

var SPORTS = [...]string{"soccer", "basketball", "football", "baseball"}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context) {
	tableName := os.Getenv("TABLE_NAME")

	marketplaceDbItems := marketplace.GetMarketplaceItems(tableName)

	fmt.Printf("%d is len", len(marketplaceDbItems))

	marketplaceCache := make(map[string]bool)

	for _, item := range marketplaceDbItems {
		marketplaceCache[item.SortKey] = true
	}

	espnResponseChannel := make(chan espnResponse, 4)

	for _, sport := range SPORTS {
		go getEspnData(sport, espnResponseChannel)
	}

	events := make([]marketplace.MarketplaceItem, 0, 25)

	for range SPORTS {
		response := <-espnResponseChannel

		for _, sport := range response.Sports {
			for _, league := range sport.Leagues {
				for _, event := range league.Events {

					awayCompetitor := event.Competitors[0]
					homeCompetitor := event.Competitors[1]

					if !marketplaceCache[marketplace.BuildMarketplaceDynamoId(sport.Name, league.Name, event.Date, awayCompetitor.Name, homeCompetitor.Name)] && event.Date.After(time.Now()) && event.Odds.Details != "" {

						events = append(events, marketplace.MarketplaceItem{Name: event.Name, AwayCompetitor: awayCompetitor.Name,
							HomeCompetitor: homeCompetitor.Name, Date: event.Date, Sport: sport.Name, League: league.Name,
							AwayLogo: awayCompetitor.Logo, HomeLogo: homeCompetitor.Logo, Spread: event.Odds.Details})
					}
				}
			}
		}
	}

	fmt.Printf("saving %d events\n", len(events))
	var waitGroup sync.WaitGroup

	for i := 0; i < len(events); i += 25 {
		waitGroup.Add(1)
		marketplace.WriteMarketplaceItems(events[i:min(i+25, len(events))], tableName, &waitGroup)
	}

	waitGroup.Wait()
}
