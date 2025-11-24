package fantasy

// PlayerStats represents real-world performance
type PlayerStats struct {
	Runs        int
	BallsFaced  int
	Fours       int
	Sixes       int
	Wickets     int
	Maidens     int
	OversBowled float64
	RunsConceded int
	Catches     int
	Stumpings   int
	RunOuts     int
	Duck        bool
}

// FantasyScorer calculates points
type FantasyScorer struct{}

// NewFantasyScorer creates a new scorer
func NewFantasyScorer() *FantasyScorer {
	return &FantasyScorer{}
}

// CalculatePoints computes points for a player based on stats
func (s *FantasyScorer) CalculatePoints(stats PlayerStats, role string) float64 {
	points := 0.0

	// 1. Batting Points
	points += float64(stats.Runs) // 1 point per run
	points += float64(stats.Fours) * 1.0 // Bonus for 4
	points += float64(stats.Sixes) * 2.0 // Bonus for 6

	// Milestones
	if stats.Runs >= 100 {
		points += 16.0
	} else if stats.Runs >= 50 {
		points += 8.0
	} else if stats.Runs >= 30 {
		points += 4.0 // T20 specific usually
	}

	// Duck penalty (only for batsmen, wicket-keepers, and all-rounders)
	if stats.Duck && (role == RoleBatsman || role == RoleWicketKeeper || role == RoleAllRounder) {
		points -= 2.0
	}

	// 2. Bowling Points
	points += float64(stats.Wickets) * 25.0
	points += float64(stats.Maidens) * 12.0 // Bonus for maiden over

	// LBW/Bowled bonus (not tracked in simple stats, assume included in Wickets)

	// Milestones
	if stats.Wickets >= 5 {
		points += 16.0
	} else if stats.Wickets >= 4 {
		points += 8.0
	} else if stats.Wickets >= 3 {
		points += 4.0
	}

	// Economy Rate Bonus (T20)
	if stats.OversBowled >= 2.0 {
		economy := float64(stats.RunsConceded) / stats.OversBowled
		if economy < 5.0 {
			points += 6.0
		} else if economy < 6.0 {
			points += 4.0
		} else if economy >= 12.0 {
			points -= 6.0
		} else if economy >= 10.0 {
			points -= 4.0
		}
	}

	// Strike Rate Bonus (Batting)
	if stats.BallsFaced >= 10 {
		strikeRate := (float64(stats.Runs) / float64(stats.BallsFaced)) * 100.0
		if strikeRate > 170.0 {
			points += 6.0
		} else if strikeRate > 150.0 {
			points += 4.0
		} else if strikeRate < 50.0 {
			points -= 6.0
		} else if strikeRate < 60.0 {
			points -= 4.0
		}
	}

	// 3. Fielding Points
	points += float64(stats.Catches) * 8.0
	points += float64(stats.Stumpings) * 12.0
	points += float64(stats.RunOuts) * 6.0

	return points
}

// CalculateTeamPoints computes total points for a user team
func (s *FantasyScorer) CalculateTeamPoints(team *FantasyTeam, matchStats map[string]PlayerStats) float64 {
	totalPoints := 0.0

	for i, player := range team.Players {
		stats, ok := matchStats[player.PlayerID]
		if !ok {
			continue
		}

		points := s.CalculatePoints(stats, player.Role)

		// Apply multipliers
		if player.PlayerID == team.CaptainID {
			points *= 2.0
		} else if player.PlayerID == team.ViceCaptainID {
			points *= 1.5
		}

		// Update individual player points in team struct (if we wanted to persist it)
		// For now just summing up
		totalPoints += points

		// Hack to update the struct in place (not ideal but works for now)
		// In real app, we'd return a new struct or map
		// team.Players[i].Points = points (FantasyPlayer struct doesn't have Points field yet, let's assume it does or ignore)
		_ = i
	}

	team.TotalPoints = totalPoints
	return totalPoints
}
