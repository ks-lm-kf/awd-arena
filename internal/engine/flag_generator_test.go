package engine

import (
	"fmt"
	"regexp"
	"testing"
)

func TestFlagGenerator_Generate(t *testing.T) {
	fg := NewFlagGenerator()

	tests := []struct {
		name     string
		gameID   int
		round    int
		teamID   int64
	}{
		{"basic", 1, 1, 100},
		{"large_ids", 999, 99, 99999},
		{"zero_round", 5, 0, 1},
	}

	flagPattern := regexp.MustCompile(`^flag\{\d+_\d+_\d+_[a-f0-9]{32}\}$`)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, err := fg.Generate(tt.gameID, tt.round, tt.teamID)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}

			// Check format
			if !flagPattern.MatchString(flag) {
				t.Errorf("Generate() flag = %v, doesn't match pattern", flag)
			}

			// Check uniqueness (generate 100 flags and check they're all different)
			flags := make(map[string]bool)
			flags[flag] = true
			for i := 0; i < 100; i++ {
				f, err := fg.Generate(tt.gameID, tt.round, tt.teamID)
				if err != nil {
					t.Fatalf("Generate() error = %v", err)
				}
				if flags[f] {
					t.Errorf("Generate() produced duplicate flag: %v", f)
				}
				flags[f] = true
			}
		})
	}
}

func TestFlagGenerator_GenerateBatch(t *testing.T) {
	fg := NewFlagGenerator()

	teamIDs := []int64{1, 2, 3, 4, 5}
	gameID := 1
	round := 1

	flags, err := fg.GenerateBatch(gameID, round, teamIDs)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	// Check all teams have flags
	if len(flags) != len(teamIDs) {
		t.Errorf("GenerateBatch() got %v flags, want %v", len(flags), len(teamIDs))
	}

	// Check all flags are unique
	seen := make(map[string]bool)
	for teamID, flag := range flags {
		if seen[flag] {
			t.Errorf("GenerateBatch() produced duplicate flag: %v", flag)
		}
		seen[flag] = true

		// Check teamID is in the flag
		if !regexp.MustCompile(fmt.Sprintf("flag\\{%d_%d_%d_", gameID, round, teamID)).MatchString(flag) {
			t.Errorf("GenerateBatch() flag %v doesn't contain correct IDs", flag)
		}
	}
}

func TestFlagGenerator_Randomness(t *testing.T) {
	fg := NewFlagGenerator()

	// Generate 1000 flags and check randomness quality
	flags := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		flag, err := fg.Generate(1, 1, 1)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		flags[flag] = true
	}

	// All should be unique
	if len(flags) != 1000 {
		t.Errorf("Generate() produced %d unique flags out of 1000", len(flags))
	}
}
