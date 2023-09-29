package main

import (
	"context"
	"net/http"
	"testing"

	"sammy.link/bet"
	"sammy.link/database"
	"sammy.link/espn"
	"sammy.link/league"
	"sammy.link/outcome"
)

func TestGetBetKinds(t *testing.T) {
	dummySlice := make([]bet.Bet, 0, 2)
	dummySlice = append(dummySlice, bet.Bet{
		Kind: "NFL",
	})
	dummySlice = append(dummySlice, bet.Bet{
		Kind: "CFB",
	})

	result := getBetKinds(dummySlice)

	if len(result) != 2 || ((result[0] != "NFL" || result[1] != "CFB") && (result[1] != "NFL" || result[0] != "CFB")) {
		t.Fatalf("Should return []{'NFL','CFB'} but got %s", result)
	}
}

func TestGetSpreads(t *testing.T) {
	slice := make([]bet.Bet, 0, 2)

	slice = append(slice, bet.Bet{
		AwayTeam: "Away",
		HomeTeam: "Home",
		Spread:   "Spread",
	})

	spreads := getSpreads(slice)

	if len(spreads) != 1 {
		t.Fatalf("should return 1 spread")
	}
}

type MockOutcomeService struct {
}

// func (s *MockOutcomeService) GetDynamoItem(o outcome.OutcomeItem) outcome.DynamoItem {
// 	return outcome.DynamoItem{}
// }

// func (s *MockOutcomeService) Write(ctx context.Context, outcomes []outcome.OutcomeItem) {
// }

// type MockBetService struct {
// }

// func (s *MockBetService) GetBetsByEventDate(ctx context.Context, date string) []bet.BetDynamoItem {
// 	return []bet.BetDynamoItem{
// 		{
// 			Id:      "BET|20230831",
// 			SortKey: "CFB|Florida|Utah|sam@sam.com|greg@greg.com",
// 		},
// 	}
// }

func TestHandler(t *testing.T) {
	ctx := context.TODO()
	handler(ctx, outcome.NewService(database.GetDatabaseService[outcome.OutcomeDynamoItem, outcome.OutcomeItem](ctx)),
		bet.NewService(database.GetDatabaseService[bet.BetDynamoItem, bet.Bet](ctx)), espn.NewService(http.Client{}),
		league.NewService(database.GetDatabaseService[league.LeagueDynamoItem, league.LeagueItem](ctx),
			database.GetDatabaseService[league.UserInLeagueDynamoItem, league.UserInLeagueItem](ctx)))
}
