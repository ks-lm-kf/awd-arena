package engine

import (
	"testing"
)

func TestFlagGenerator_Generate_Uniqueness(t *testing.T) {
	fg := NewFlagGenerator()

	// Generate 1000 flags for the same team/round and ensure uniqueness
	flags := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		flag, err := fg.Generate(1, 1, 100)
		if err != nil {
			t.Fatalf("Generate() error = %v", err)
		}
		if flags[flag] {
			t.Errorf("Duplicate flag generated: %s", flag)
		}
		flags[flag] = true
	}
}

func TestFlagGenerator_Generate_LargeBatch(t *testing.T) {
	fg := NewFlagGenerator()

	// Generate flags for 100 teams
	teamIDs := make([]int64, 100)
	for i := 0; i < 100; i++ {
		teamIDs[i] = int64(i + 1)
	}

	flags, err := fg.GenerateBatch(1, 1, teamIDs)
	if err != nil {
		t.Fatalf("GenerateBatch() error = %v", err)
	}

	if len(flags) != 100 {
		t.Errorf("GenerateBatch() got %d flags, want 100", len(flags))
	}

	// Check uniqueness
	seen := make(map[string]bool)
	for _, flag := range flags {
		if seen[flag] {
			t.Errorf("Duplicate flag in batch: %s", flag)
		}
		seen[flag] = true
	}
}

func TestFlagGenerator_Generate_EdgeCases(t *testing.T) {
	fg := NewFlagGenerator()

	tests := []struct {
		name   string
		gameID int
		round  int
		teamID int64
	}{
		{"zero_game_id", 0, 1, 100},
		{"large_game_id", 999999999, 1, 100},
		{"negative_round", 1, -1, 100},
		{"zero_team_id", 1, 1, 0},
		{"large_team_id", 1, 1, 999999999999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag, err := fg.Generate(tt.gameID, tt.round, tt.teamID)
			if err != nil {
				t.Fatalf("Generate() error = %v", err)
			}
			if flag == "" {
				t.Error("Generate() returned empty flag")
			}
		})
	}
}

func TestFlagGenerator_NewFlagGenerator_Defaults(t *testing.T) {
	fg := NewFlagGenerator()
	if fg.randomLength != 16 {
		t.Errorf("randomLength = %d, want 16", fg.randomLength)
	}
}

func TestFlagWriter_WriteFlagBatch_EmptyFlags(t *testing.T) {
	writer := NewFlagWriter(nil)
	err := writer.WriteFlagBatch(nil, map[string]string{})
	if err != nil {
		t.Errorf("WriteFlagBatch with empty map should not error: %v", err)
	}
}

func TestFlagWriter_WriteFlagBatch_SingleFlag(t *testing.T) {
	writer := NewFlagWriter(nil)
	flags := map[string]string{
		"container1": "flag{test}",
	}
	err := writer.WriteFlagBatch(nil, flags)
	// Should fail with nil client
	if err == nil {
		t.Error("WriteFlagBatch with nil client should return error")
	}
}
