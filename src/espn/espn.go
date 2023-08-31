package espn

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type EspnResponse struct {
	Sports []EspnSport `json:"sports"`
}

type EspnSport struct {
	Name    string       `json:"name"`
	Leagues []EspnLeague `json:"leagues"`
}

type EspnLeague struct {
	Name         string      `json:"name"`
	Abbreviation string      `json:"abbreviation"`
	Events       []EspnEvent `json:"events"`
}

type EspnEvent struct {
	Id          string           `json:"id"`
	Name        string           `json:"name"`
	ShortName   string           `json:"shortName"`
	Date        time.Time        `json:"date"`
	Odds        EspnOdds         `json:"odds"`
	Status      string           `json:"status"`
	Competitors []EspnCompetitor `json:"competitors"`
	Week        int              `json:"week"`
}

type EspnCompetitor struct {
	Name         string `json:"name"`
	HomeAway     string `json:"homeAway"`
	Winner       bool   `json:"winner"`
	Score        string `json:"score"`
	Abbreviation string `json:"abbreviation"`
}

type EspnOdds struct {
	Details string `json:"details"`
}

type Kind struct {
	Sport  string
	League string
}

func ConvertKind(kind string) Kind {
	if kind == "CFB" {
		return Kind{
			Sport:  "football",
			League: "college-football",
		}
	}
	return Kind{
		Sport:  "football",
		League: "nfl",
	}
}

// time.Now().AddDate(0, 0, 7)
func GetEspnData(sport string, league string, channel chan EspnResponse, date time.Time, secondDate time.Time) {
	const format = "20060102"

	resp, err := http.Get(fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=%s&league=%s&dates=%s-%s",
		sport, league, date.Format(format), secondDate.Format(format)))

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	var response EspnResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		panic(err)
	}

	channel <- response
}
