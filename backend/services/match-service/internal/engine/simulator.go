package engine

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/playkaro/match-service/internal/cache"
)

// MatchSimulator simulates a live cricket match
type MatchSimulator struct {
	MatchID     string
	TeamA       string
	TeamB       string
	ScoreA      int
	WicketsA    int
	OversA      float64
	Target      int
	IsChasing   bool
	OddsA       float64
	OddsB       float64
	Cache       *cache.MatchCache
	StopChan    chan bool
}

func NewMatchSimulator(matchID, teamA, teamB string, cache *cache.MatchCache) *MatchSimulator {
	return &MatchSimulator{
		MatchID:   matchID,
		TeamA:     teamA,
		TeamB:     teamB,
		OddsA:     1.90,
		OddsB:     1.90,
		Cache:     cache,
		StopChan:  make(chan bool),
	}
}

func (s *MatchSimulator) Start() {
	ticker := time.NewTicker(2 * time.Second) // Update every 2 seconds

	go func() {
		for {
			select {
			case <-ticker.C:
				s.simulateBall()
				s.broadcastUpdate()
			case <-s.StopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

func (s *MatchSimulator) Stop() {
	s.StopChan <- true
}

func (s *MatchSimulator) simulateBall() {
	// 1. Generate Event (0, 1, 2, 3, 4, 6, W)
	events := []string{"0", "1", "1", "2", "4", "6", "W", "0", "1"}
	event := events[rand.Intn(len(events))]

	// 2. Update Score
	if event == "W" {
		s.WicketsA++
		// Wicket falls -> Odds shift dramatically against batting team
		s.OddsA += 0.5
		s.OddsB -= 0.2
	} else {
		runs := 0
		fmt.Sscanf(event, "%d", &runs)
		s.ScoreA += runs

		// Boundaries -> Odds shift slightly in favor of batting team
		if runs >= 4 {
			s.OddsA -= 0.05
			s.OddsB += 0.05
		}
	}

	// Update Overs
	// Simple logic: 0.1, 0.2... 0.5 -> 1.0
	// For demo, just incrementing balls count conceptually

	// Normalize Odds (Min 1.01, Max 100.0)
	if s.OddsA < 1.01 { s.OddsA = 1.01 }
	if s.OddsB < 1.01 { s.OddsB = 1.01 }
}

func (s *MatchSimulator) broadcastUpdate() {
	s.Cache.PublishOddsUpdate(context.Background(), s.MatchID, s.OddsA, s.OddsB, 0)
}
