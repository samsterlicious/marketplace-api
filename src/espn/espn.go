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
	Link        string           `json:"link"`
}

type EspnCompetitor struct {
	Name         string `json:"name"`
	HomeAway     string `json:"homeAway"`
	Winner       bool   `json:"winner"`
	Score        string `json:"score"`
	Abbreviation string `json:"abbreviation"`
	Record       string `json:"record"`
}

type EspnOdds struct {
	Details string `json:"details"`
}

type Kind struct {
	Sport  string
	League string
}

type Service interface {
	ConvertKind(kind string) Kind
	GetEspnData(sport string, league string, channel chan EspnResponse, date time.Time, secondDate time.Time)
	GetEspnEvent(sport string, league string, eventId string) (EspnResponse, error)
}

type EspnService struct {
	client http.Client
}

func NewService(client http.Client) Service {
	return &EspnService{
		client: client,
	}
}

func (s *EspnService) ConvertKind(kind string) Kind {
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

func (s *EspnService) GetEspnEvent(sport string, league string, eventId string) (EspnResponse, error) {
	request, _ := http.NewRequest("GET", fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=%s&league=%s&event=%s",
		sport, league, eventId), nil)

	resp, err := s.client.Do(request)

	var response EspnResponse

	if err != nil {
		return response, err
	}

	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		return response, err
	}

	return response, nil
}

func (s *EspnService) GetEspnData(sport string, league string, channel chan EspnResponse, date time.Time, secondDate time.Time) {
	const format = "20060102"

	request, err := http.NewRequest("GET", fmt.Sprintf("https://site.web.api.espn.com/apis/v2/scoreboard/header?sport=%s&league=%s&dates=%s-%s",
		sport, league, date.Format(format), secondDate.Format(format)), nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	resp, err := s.client.Do(request)

	if err != nil {
		fmt.Println(err.Error())
	}

	defer resp.Body.Close()

	var response EspnResponse
	err = json.NewDecoder(resp.Body).Decode(&response)

	if err != nil {
		fmt.Println(err.Error())
	}
	channel <- response
}
