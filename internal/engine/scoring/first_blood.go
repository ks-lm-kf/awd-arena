package scoring

import (
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

// FirstBloodDetector tracks and rewards the first team to capture each flag.
type FirstBloodDetector struct {
	mu sync.RWMutex

	// firstBlood records which team first captured each flag
	// key: flag hash/identifier, value: FirstBloodRecord
	records map[string]*FirstBloodRecord
}

// FirstBloodRecord stores information about a first blood capture.
type FirstBloodRecord struct {
	FlagID      string // Flag identifier
	TeamID      int64  // Team that got first blood
	Round       int    // Round when it happened
	Timestamp   int64  // Unix timestamp
	BonusPoints float64 // Bonus points awarded
}

// NewFirstBloodDetector creates a new first blood detector.
func NewFirstBloodDetector() *FirstBloodDetector {
	return &FirstBloodDetector{
		records: make(map[string]*FirstBloodRecord),
	}
}

// CheckAndRecord checks if this is the first submission for a flag.
// Returns true if this is first blood, false otherwise.
// Thread-safe and race-condition resistant.
func (d *FirstBloodDetector) CheckAndRecord(flagID string, teamID int64, round int) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check if flag already has a first blood record
	if _, exists := d.records[flagID]; exists {
		logger.Debug("flag already has first blood",
			"flag", flagID,
			"team", teamID,
		)
		return false
	}

	// Record first blood
	d.records[flagID] = &FirstBloodRecord{
		FlagID:    flagID,
		TeamID:    teamID,
		Round:     round,
		Timestamp: getCurrentTimestamp(),
	}

	logger.Info("first blood recorded",
		"flag", flagID,
		"team", teamID,
		"round", round,
	)

	return true
}

// HasFirstBlood checks if a flag has been captured for first blood.
func (d *FirstBloodDetector) HasFirstBlood(flagID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	_, exists := d.records[flagID]
	return exists
}

// GetFirstBloodRecord returns the first blood record for a flag.
func (d *FirstBloodDetector) GetFirstBloodRecord(flagID string) *FirstBloodRecord {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.records[flagID]
}

// GetTeamFirstBloods returns all first blood records for a team.
func (d *FirstBloodDetector) GetTeamFirstBloods(teamID int64) []*FirstBloodRecord {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var teamRecords []*FirstBloodRecord
	for _, record := range d.records {
		if record.TeamID == teamID {
			teamRecords = append(teamRecords, record)
		}
	}
	return teamRecords
}

// GetAllFirstBloods returns all first blood records.
func (d *FirstBloodDetector) GetAllFirstBloods() map[string]*FirstBloodRecord {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]*FirstBloodRecord)
	for k, v := range d.records {
		recordCopy := *v
		result[k] = &recordCopy
	}
	return result
}

// Clear resets all first blood records.
// Should be called when starting a new game.
func (d *FirstBloodDetector) Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.records = make(map[string]*FirstBloodRecord)
	logger.Info("first blood records cleared")
}

// LoadFromRecords loads first blood records from persistence.
// Used when recovering state after a restart.
func (d *FirstBloodDetector) LoadFromRecords(records []*FirstBloodRecord) {
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, record := range records {
		d.records[record.FlagID] = record
	}
	
	logger.Info("first blood records loaded", "count", len(records))
}

// getCurrentTimestamp returns current Unix timestamp.
// Extracted for testability.
func getCurrentTimestamp() int64 {
	return now().Unix()
}

// Time provider for testability
var now = func() interface{ Unix() int64 } {
	return time.Now()
}

// currentTimeUnix is deprecated, use now() instead.
func currentTimeUnix() int64 {
	return time.Now().Unix()
}
