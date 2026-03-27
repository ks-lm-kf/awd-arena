package scoring

import (
	"sync"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

// DefenseCalculator calculates defense scores for teams.
// Teams that successfully defend their flags earn points.
type DefenseCalculator struct {
	mu sync.RWMutex

	defenseValuePerRound float64 // Points per round for successful defense
	
	// Track how many times each team was breached per round
	// key: round, value: map[teamID]breachCount
	breachHistory map[int]map[int64]int
}

// NewDefenseCalculator creates a new defense calculator.
func NewDefenseCalculator(defenseValue float64) *DefenseCalculator {
	return &DefenseCalculator{
		defenseValuePerRound: defenseValue,
		breachHistory:        make(map[int]map[int64]int),
	}
}

// RecordBreach records that a team was breached in a specific round.
func (d *DefenseCalculator) RecordBreach(round int, teamID int64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.breachHistory[round] == nil {
		d.breachHistory[round] = make(map[int64]int)
	}
	
	d.breachHistory[round][teamID]++
	
	logger.Debug("breach recorded",
		"round", round,
		"team", teamID,
		"count", d.breachHistory[round][teamID],
	)
}

// CalculateDefenseBonus calculates defense bonus for a team.
// Formula: (totalRounds - timesBreached) × defenseValue
// Teams that were never breached get maximum bonus.
func (d *DefenseCalculator) CalculateDefenseBonus(round int, totalRounds int, timesBreached int) float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Teams that were breached lose defense points proportionally
	// More breaches = less defense bonus
	successfulDefenses := totalRounds - timesBreached
	
	// Defense bonus scales with how many times you defended successfully
	bonus := float64(successfulDefenses) * d.defenseValuePerRound / float64(totalRounds)
	
	// Ensure minimum defense score is 0 (not negative)
	if bonus < 0 {
		bonus = 0
	}
	
	return bonus
}

// GetTeamBreaches returns total breaches for a team across all rounds.
func (d *DefenseCalculator) GetTeamBreaches(teamID int64) int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	total := 0
	for _, roundBreaches := range d.breachHistory {
		total += roundBreaches[teamID]
	}
	return total
}

// GetRoundBreaches returns breach count for a specific round.
func (d *DefenseCalculator) GetRoundBreaches(round int) map[int64]int {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if d.breachHistory[round] == nil {
		return make(map[int64]int)
	}

	// Return a copy
	result := make(map[int64]int)
	for k, v := range d.breachHistory[round] {
		result[k] = v
	}
	return result
}

// Clear clears all breach history.
func (d *DefenseCalculator) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.breachHistory = make(map[int]map[int64]int)
	logger.Info("defense breach history cleared")
}

// SetDefenseValue updates the defense value per round.
func (d *DefenseCalculator) SetDefenseValue(value float64) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.defenseValuePerRound = value
}

// GetDefenseValue returns current defense value.
func (d *DefenseCalculator) GetDefenseValue() float64 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.defenseValuePerRound
}

// DefenseStats provides statistics about defense performance.
type DefenseStats struct {
	TeamID             int64
	TotalBreaches      int
	SuccessfulDefenses int
	DefenseScore       float64
}

// GetDefenseStats returns defense statistics for all teams.
func (d *DefenseCalculator) GetDefenseStats(totalRounds int, teamIDs []int64) []DefenseStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make([]DefenseStats, 0, len(teamIDs))
	
	for _, teamID := range teamIDs {
		breaches := 0
		for _, roundBreaches := range d.breachHistory {
			breaches += roundBreaches[teamID]
		}
		
		successfulDefenses := totalRounds - breaches
		if successfulDefenses < 0 {
			successfulDefenses = 0
		}
		
		bonus := float64(successfulDefenses) * d.defenseValuePerRound / float64(totalRounds)
		
		stats = append(stats, DefenseStats{
			TeamID:             teamID,
			TotalBreaches:      breaches,
			SuccessfulDefenses: successfulDefenses,
			DefenseScore:       bonus,
		})
	}
	
	return stats
}
