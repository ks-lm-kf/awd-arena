package security

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// AttackLogEntry records an attack attempt in memory.
type AttackLogEntry struct {
	ID           string    `json:"id"`
	Timestamp    time.Time `json:"timestamp"`
	GameID       string    `json:"game_id"`
	AttackerTeam string    `json:"attacker_team"`
	TargetTeam   string    `json:"target_team,omitempty"`
	AttackType   string    `json:"attack_type"`   // sql_injection, xss, command_injection, path_traversal, brute_force, other
	Severity     string    `json:"severity"`       // low, medium, high, critical
	SourceIP     string    `json:"source_ip"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Payload      string    `json:"payload,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	RuleMatched  string    `json:"rule_matched,omitempty"`
	Blocked      bool      `json:"blocked"`
}

// AttackLogStore manages attack logs in memory.
type AttackLogStore struct {
	mu      sync.RWMutex
	logs    []AttackLogEntry
	maxSize int
}

// NewAttackLogStore creates a new attack log store.
func NewAttackLogStore(maxSize int) *AttackLogStore {
	return &AttackLogStore{
		logs:    make([]AttackLogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add records a new attack log entry.
func (s *AttackLogStore) Add(entry AttackLogEntry) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.ID == "" {
		entry.ID = uuid.New().String()
	}
	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now()
	}
	s.logs = append(s.logs, entry)
	if len(s.logs) > s.maxSize {
		s.logs = s.logs[len(s.logs)-s.maxSize:]
	}
}

// Query returns attack logs with optional filters.
func (s *AttackLogStore) Query(gameID, teamID, attackType, severity string, limit, offset int) []AttackLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []AttackLogEntry
	// Iterate in reverse (newest first)
	for i := len(s.logs) - 1; i >= 0; i-- {
		entry := s.logs[i]
		if gameID != "" && entry.GameID != gameID {
			continue
		}
		if teamID != "" && entry.AttackerTeam != teamID {
			continue
		}
		if attackType != "" && entry.AttackType != attackType {
			continue
		}
		if severity != "" && entry.Severity != severity {
			continue
		}
		result = append(result, entry)
		if len(result) >= limit {
			break
		}
	}
	// Apply offset
	if offset >= len(result) {
		return nil
	}
	return result[offset:]
}

// Count returns total count for a filter.
func (s *AttackLogStore) Count(gameID, attackType, severity string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	count := 0
	for _, entry := range s.logs {
		if gameID != "" && entry.GameID != gameID {
			continue
		}
		if attackType != "" && entry.AttackType != attackType {
			continue
		}
		if severity != "" && entry.Severity != severity {
			continue
		}
		count++
	}
	return count
}

// GetByID returns a specific attack log by ID.
func (s *AttackLogStore) GetByID(id string) *AttackLogEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, entry := range s.logs {
		if entry.ID == id {
			return &entry
		}
	}
	return nil
}
