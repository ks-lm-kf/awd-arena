package engine

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/awd-platform/awd-arena/pkg/logger"
)

// FlagGenerator generates secure flags for AWD rounds.
type FlagGenerator struct {
	randomLength int // length of random part in bytes
}

// NewFlagGenerator creates a new flag generator.
func NewFlagGenerator() *FlagGenerator {
	return &FlagGenerator{
		randomLength: 16, // 16 bytes = 32 hex chars
	}
}

// Generate generates a flag with the format: flag{gameid_round_teamid_random}
func (fg *FlagGenerator) Generate(gameID, round int, teamID int64) (string, error) {
	random, err := fg.generateRandom()
	if err != nil {
		logger.Error("failed to generate random part", "error", err)
		return "", fmt.Errorf("failed to generate random: %w", err)
	}

	flag := fmt.Sprintf("flag{%d_%d_%d_%s}", gameID, round, teamID, random)
	logger.Debug("generated flag", "game_id", gameID, "round", round, "team_id", teamID)

	return flag, nil
}

// generateRandom generates a cryptographically secure random hex string.
func (fg *FlagGenerator) generateRandom() (string, error) {
	bytes := make([]byte, fg.randomLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("crypto/rand failed: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateBatch generates multiple flags for a round.
func (fg *FlagGenerator) GenerateBatch(gameID, round int, teamIDs []int64) (map[int64]string, error) {
	flags := make(map[int64]string, len(teamIDs))

	for _, teamID := range teamIDs {
		flag, err := fg.Generate(gameID, round, teamID)
		if err != nil {
			return nil, fmt.Errorf("failed to generate flag for team %d: %w", teamID, err)
		}
		flags[teamID] = flag
	}

	logger.Info("generated batch of flags", "count", len(flags), "game_id", gameID, "round", round)
	return flags, nil
}
