package scoring

import (
	"testing"
)

func TestDefenseCalculator_NewDefenseCalculator(t *testing.T) {
	dc := NewDefenseCalculator(50.0)
	if dc == nil {
		t.Fatal("NewDefenseCalculator returned nil")
	}
	if dc.GetDefenseValue() != 50.0 {
		t.Errorf("Expected defense value 50, got %f", dc.GetDefenseValue())
	}
}

func TestDefenseCalculator_RecordBreach(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	// Record breaches
	dc.RecordBreach(1, 1)
	dc.RecordBreach(1, 1)
	dc.RecordBreach(1, 2)
	dc.RecordBreach(2, 1)

	// Check round 1 breaches
	round1Breaches := dc.GetRoundBreaches(1)
	if round1Breaches[1] != 2 {
		t.Errorf("Team 1 should have 2 breaches in round 1, got %d", round1Breaches[1])
	}
	if round1Breaches[2] != 1 {
		t.Errorf("Team 2 should have 1 breach in round 1, got %d", round1Breaches[2])
	}

	// Check total breaches for team 1
	totalBreaches := dc.GetTeamBreaches(1)
	if totalBreaches != 3 {
		t.Errorf("Team 1 should have 3 total breaches, got %d", totalBreaches)
	}
}

func TestDefenseCalculator_CalculateDefenseBonus(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	tests := []struct {
		name          string
		round         int
		totalRounds   int
		timesBreached int
		wantMin       float64
		wantMax       float64
	}{
		{
			name:          "No breaches",
			round:         1,
			totalRounds:   10,
			timesBreached: 0,
			wantMin:       0,
			wantMax:       50.0,
		},
		{
			name:          "Some breaches",
			round:         1,
			totalRounds:   10,
			timesBreached: 3,
			wantMin:       0,
			wantMax:       35.0,
		},
		{
			name:          "All breached",
			round:         1,
			totalRounds:   10,
			timesBreached: 10,
			wantMin:       0,
			wantMax:       0,
		},
		{
			name:          "Over breached (should not be negative)",
			round:         1,
			totalRounds:   10,
			timesBreached: 15,
			wantMin:       0,
			wantMax:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := dc.CalculateDefenseBonus(tt.round, tt.totalRounds, tt.timesBreached)
			if bonus < tt.wantMin || bonus > tt.wantMax {
				t.Errorf("CalculateDefenseBonus() = %f, want between %f and %f", bonus, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestDefenseCalculator_GetRoundBreaches(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	// Record breaches in multiple rounds
	dc.RecordBreach(1, 1)
	dc.RecordBreach(1, 2)
	dc.RecordBreach(2, 1)
	dc.RecordBreach(2, 3)

	// Check round 1
	round1 := dc.GetRoundBreaches(1)
	if len(round1) != 2 {
		t.Errorf("Round 1 should have 2 teams breached, got %d", len(round1))
	}

	// Check round 2
	round2 := dc.GetRoundBreaches(2)
	if len(round2) != 2 {
		t.Errorf("Round 2 should have 2 teams breached, got %d", len(round2))
	}

	// Check non-existent round
	round99 := dc.GetRoundBreaches(99)
	if len(round99) != 0 {
		t.Errorf("Non-existent round should have 0 breaches, got %d", len(round99))
	}
}

func TestDefenseCalculator_GetTeamBreaches(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	// Team 1: 3 breaches across 2 rounds
	dc.RecordBreach(1, 1)
	dc.RecordBreach(2, 1)
	dc.RecordBreach(3, 1)

	// Team 2: 1 breach
	dc.RecordBreach(1, 2)

	// Team 3: 0 breaches

	if dc.GetTeamBreaches(1) != 3 {
		t.Errorf("Team 1 should have 3 breaches, got %d", dc.GetTeamBreaches(1))
	}
	if dc.GetTeamBreaches(2) != 1 {
		t.Errorf("Team 2 should have 1 breach, got %d", dc.GetTeamBreaches(2))
	}
	if dc.GetTeamBreaches(3) != 0 {
		t.Errorf("Team 3 should have 0 breaches, got %d", dc.GetTeamBreaches(3))
	}
}

func TestDefenseCalculator_Clear(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	// Record some breaches
	dc.RecordBreach(1, 1)
	dc.RecordBreach(2, 2)

	// Clear
	dc.Clear()

	// Check cleared
	if dc.GetTeamBreaches(1) != 0 {
		t.Error("Team 1 breaches should be cleared")
	}
	if dc.GetTeamBreaches(2) != 0 {
		t.Error("Team 2 breaches should be cleared")
	}
	if len(dc.GetRoundBreaches(1)) != 0 {
		t.Error("Round 1 breaches should be cleared")
	}
}

func TestDefenseCalculator_SetDefenseValue(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	// Update defense value
	dc.SetDefenseValue(75.0)

	if dc.GetDefenseValue() != 75.0 {
		t.Errorf("Expected defense value 75, got %f", dc.GetDefenseValue())
	}
}

func TestDefenseCalculator_GetDefenseStats(t *testing.T) {
	dc := NewDefenseCalculator(50.0)
	teamIDs := []int64{1, 2, 3}
	totalRounds := 10

	// Record breaches
	dc.RecordBreach(1, 1)
	dc.RecordBreach(2, 1)
	dc.RecordBreach(3, 2)

	stats := dc.GetDefenseStats(totalRounds, teamIDs)

	if len(stats) != 3 {
		t.Errorf("Expected 3 team stats, got %d", len(stats))
	}

	// Team 1: 2 breaches
	team1Stats := stats[0]
	if team1Stats.TotalBreaches != 2 {
		t.Errorf("Team 1 should have 2 breaches, got %d", team1Stats.TotalBreaches)
	}
	if team1Stats.SuccessfulDefenses != 8 {
		t.Errorf("Team 1 should have 8 successful defenses, got %d", team1Stats.SuccessfulDefenses)
	}

	// Team 2: 1 breach
	team2Stats := stats[1]
	if team2Stats.TotalBreaches != 1 {
		t.Errorf("Team 2 should have 1 breach, got %d", team2Stats.TotalBreaches)
	}

	// Team 3: 0 breaches
	team3Stats := stats[2]
	if team3Stats.TotalBreaches != 0 {
		t.Errorf("Team 3 should have 0 breaches, got %d", team3Stats.TotalBreaches)
	}
}

func TestDefenseCalculator_ConcurrentBreaches(t *testing.T) {
	dc := NewDefenseCalculator(50.0)

	done := make(chan bool, 100)

	// Concurrent breach recording
	for i := 0; i < 100; i++ {
		go func(idx int) {
			round := (idx % 10) + 1
			team := int64((idx % 5) + 1)
			dc.RecordBreach(round, team)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Verify total breaches
	totalBreaches := 0
	for team := 1; team <= 5; team++ {
		totalBreaches += dc.GetTeamBreaches(int64(team))
	}

	if totalBreaches != 100 {
		t.Errorf("Expected 100 total breaches, got %d", totalBreaches)
	}
}
