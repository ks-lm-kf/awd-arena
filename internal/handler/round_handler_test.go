package handler

import (
	"testing"
)

func TestNewRoundHandler(t *testing.T) {
	handler := NewRoundHandler()
	
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	
	if handler.managers == nil {
		t.Error("expected managers map to be initialized")
	}
}

func TestRoundInfoResponse_Validation(t *testing.T) {
	resp := RoundInfoResponse{
		GameID:         1,
		GameTitle:      "Test Game",
		CurrentRound:   1,
		TotalRounds:    10,
		Phase:          "running",
		RoundDuration:  300,
		BreakDuration:  60,
		ElapsedTime:    150,
		RemainingTime:  150,
		IsPaused:       false,
	}
	
	if resp.GameID != 1 {
		t.Errorf("expected game ID 1, got %d", resp.GameID)
	}
	
	if resp.CurrentRound != 1 {
		t.Errorf("expected current round 1, got %d", resp.CurrentRound)
	}
	
	if resp.ElapsedTime + resp.RemainingTime != resp.RoundDuration {
		t.Errorf("elapsed + remaining should equal round duration")
	}
}

func TestRoundControlRequest_Validation(t *testing.T) {
	req := RoundControlRequest{
		Action: "pause",
	}
	
	if req.Action != "pause" {
		t.Errorf("expected action pause, got %s", req.Action)
	}
	
	// Test with optional fields
	reqWithDurations := RoundControlRequest{
		Action:        "start",
		RoundDuration: intPtr(600),
		BreakDuration: intPtr(120),
	}
	
	if *reqWithDurations.RoundDuration != 600 {
		t.Errorf("expected round duration 600, got %d", *reqWithDurations.RoundDuration)
	}
}

func TestRoundControlResponse_Validation(t *testing.T) {
	resp := RoundControlResponse{
		Success: true,
		Message: "Operation completed",
		State: &RoundInfoResponse{
			GameID:       1,
			CurrentRound: 2,
		},
	}
	
	if !resp.Success {
		t.Error("expected success to be true")
	}
	
	if resp.State == nil {
		t.Error("expected state to be non-nil")
	}
}

// Helper function
func intPtr(i int) *int {
	return &i
}
