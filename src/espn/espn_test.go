package espn

import (
	"net/http"
	"testing"
	"time"
)

func TestGetEspnData(t *testing.T) {
	s := NewService(http.Client{})

	testChan := make(chan EspnResponse, 1)
	s.GetEspnData("football", "nfl", testChan, time.Now(), time.Now().AddDate(0, 0, 7))

	resp := <-testChan

	if len(resp.Sports) < 1 {
		t.Fatalf("Sports should be returned but received %+v", resp)
	}

}

func TestGetEspnEvent(t *testing.T) {
	s := NewService(http.Client{})

	resp, _ := s.GetEspnEvent("football", "nfl", "401547658")

	if len(resp.Sports[0].Leagues[0].Events) == 0 {
		t.Fatalf("Only got this back %+v", resp)
	}

	resp, _ = s.GetEspnEvent("football", "college-football", "401547658")

	if len(resp.Sports) > 0 {
		t.Fatalf("Should be zero response %+v", resp)
	}
}
