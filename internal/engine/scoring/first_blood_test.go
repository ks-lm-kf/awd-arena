package scoring

import (
	"testing"
)

func TestFirstBloodDetector_CheckAndRecord(t *testing.T) {
	detector := NewFirstBloodDetector()

	// First submission for flag1
	isFirst := detector.CheckAndRecord("flag1", 1, 1)
	if !isFirst {
		t.Error("Expected first blood for flag1, team 1")
	}

	// Second submission for same flag
	isFirst = detector.CheckAndRecord("flag1", 2, 1)
	if isFirst {
		t.Error("Should not be first blood for same flag")
	}

	// First submission for flag2
	isFirst = detector.CheckAndRecord("flag2", 2, 1)
	if !isFirst {
		t.Error("Expected first blood for flag2, team 2")
	}
}

func TestFirstBloodDetector_HasFirstBlood(t *testing.T) {
	detector := NewFirstBloodDetector()

	// Initially no first blood
	if detector.HasFirstBlood("flag1") {
		t.Error("Should not have first blood initially")
	}

	// Record first blood
	detector.CheckAndRecord("flag1", 1, 1)

	// Now should have first blood
	if !detector.HasFirstBlood("flag1") {
		t.Error("Should have first blood after recording")
	}
}

func TestFirstBloodDetector_GetFirstBloodRecord(t *testing.T) {
	detector := NewFirstBloodDetector()

	// Record first blood
	detector.CheckAndRecord("flag1", 1, 1)

	// Get record
	record := detector.GetFirstBloodRecord("flag1")
	if record == nil {
		t.Fatal("Expected record, got nil")
	}

	if record.FlagID != "flag1" {
		t.Errorf("Expected flag1, got %s", record.FlagID)
	}
	if record.TeamID != 1 {
		t.Errorf("Expected team 1, got %d", record.TeamID)
	}
	if record.Round != 1 {
		t.Errorf("Expected round 1, got %d", record.Round)
	}
}

func TestFirstBloodDetector_GetTeamFirstBloods(t *testing.T) {
	detector := NewFirstBloodDetector()

	// Team 1 gets first blood on flag1 and flag2
	detector.CheckAndRecord("flag1", 1, 1)
	detector.CheckAndRecord("flag2", 1, 1)

	// Team 2 gets first blood on flag3
	detector.CheckAndRecord("flag3", 2, 1)

	// Get team 1's first bloods
	team1Records := detector.GetTeamFirstBloods(1)
	if len(team1Records) != 2 {
		t.Errorf("Team 1 should have 2 first bloods, got %d", len(team1Records))
	}

	// Get team 2's first bloods
	team2Records := detector.GetTeamFirstBloods(2)
	if len(team2Records) != 1 {
		t.Errorf("Team 2 should have 1 first blood, got %d", len(team2Records))
	}

	// Get team 3's first bloods (none)
	team3Records := detector.GetTeamFirstBloods(3)
	if len(team3Records) != 0 {
		t.Errorf("Team 3 should have 0 first bloods, got %d", len(team3Records))
	}
}

func TestFirstBloodDetector_Clear(t *testing.T) {
	detector := NewFirstBloodDetector()

	// Record some first bloods
	detector.CheckAndRecord("flag1", 1, 1)
	detector.CheckAndRecord("flag2", 2, 1)

	// Clear
	detector.Clear()

	// Check if cleared
	if detector.HasFirstBlood("flag1") {
		t.Error("flag1 should be cleared")
	}
	if detector.HasFirstBlood("flag2") {
		t.Error("flag2 should be cleared")
	}

	// Should be able to record new first bloods
	isFirst := detector.CheckAndRecord("flag1", 3, 1)
	if !isFirst {
		t.Error("Should be first blood after clear")
	}
}

func TestFirstBloodDetector_Concurrency(t *testing.T) {
	detector := NewFirstBloodDetector()

	// Simulate concurrent submissions
	done := make(chan bool, 10)
	firstBloodCount := 0

	for i := 0; i < 10; i++ {
		go func() {
			isFirst := detector.CheckAndRecord("flag1", 1, 1)
			if isFirst {
				firstBloodCount++
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Only one should be first blood
	if firstBloodCount != 1 {
		t.Errorf("Expected 1 first blood, got %d", firstBloodCount)
	}
}
