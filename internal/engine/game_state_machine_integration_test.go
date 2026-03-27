package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
)

// TestNewGameStateMachine_NilGame tests nil game handling
func TestNewGameStateMachine_NilGame_Check(t *testing.T) {
	gsm := NewGameStateMachine(nil, nil, false)
	if gsm != nil {
		t.Error("Expected nil GameStateMachine for nil game")
	}
}

// TestGameStateMachine_InitialState_AllStates tests all initial state mappings
func TestGameStateMachine_InitialState_AllStates(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		currentPhase  string
		expectedState GameState
	}{
		{"Draft/Preparation", "draft", "preparation", StatePreparing},
		{"Active/Running", "active", "running", StateRunning},
		{"Active/Break", "active", "break", StatePaused},
		{"Finished/Finished", "finished", "finished", StateFinished},
		{"Unknown status", "unknown", "", StatePreparing},
		{"Empty status", "", "", StatePreparing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &model.Game{
				ID:           1,
				Status:       tt.status,
				CurrentPhase: tt.currentPhase,
			}
			gsm := NewGameStateMachine(game, nil, false)
			if gsm == nil {
				t.Fatal("GameStateMachine should not be nil")
			}
			if gsm.GetCurrentState() != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, gsm.GetCurrentState())
			}
		})
	}
}

// TestGameStateMachine_Transition_AllValidPaths tests all valid state transition paths
func TestGameStateMachine_Transition_AllValidPaths(t *testing.T) {
	ctx := context.Background()
	eb := &eventbus.Bus{}

	// Path 1: Preparing -> Running -> Paused -> Running -> Finished
	t.Run("Path1_Preparing_Running_Paused_Running_Finished", func(t *testing.T) {
		game := &model.Game{
			ID:           1,
			Status:       "draft",
			CurrentPhase: "preparation",
		}
		gsm := NewGameStateMachine(game, eb, false)

		// Start
		if err := gsm.Transition(ctx, EventStart); err != nil {
			t.Fatalf("Failed to start: %v", err)
		}
		if gsm.GetCurrentState() != StateRunning {
			t.Errorf("Expected %s, got %s", StateRunning, gsm.GetCurrentState())
		}

		// Pause
		if err := gsm.Transition(ctx, EventPause); err != nil {
			t.Fatalf("Failed to pause: %v", err)
		}
		if gsm.GetCurrentState() != StatePaused {
			t.Errorf("Expected %s, got %s", StatePaused, gsm.GetCurrentState())
		}

		// Resume
		if err := gsm.Transition(ctx, EventResume); err != nil {
			t.Fatalf("Failed to resume: %v", err)
		}
		if gsm.GetCurrentState() != StateRunning {
			t.Errorf("Expected %s, got %s", StateRunning, gsm.GetCurrentState())
		}

		// Finish
		if err := gsm.Transition(ctx, EventFinish); err != nil {
			t.Fatalf("Failed to finish: %v", err)
		}
		if gsm.GetCurrentState() != StateFinished {
			t.Errorf("Expected %s, got %s", StateFinished, gsm.GetCurrentState())
		}
	})

	// Path 2: Preparing -> Running -> Finished
	t.Run("Path2_Preparing_Running_Finished", func(t *testing.T) {
		game := &model.Game{
			ID:           2,
			Status:       "draft",
			CurrentPhase: "preparation",
		}
		gsm := NewGameStateMachine(game, eb, false)

		// Start
		if err := gsm.Transition(ctx, EventStart); err != nil {
			t.Fatalf("Failed to start: %v", err)
		}

		// Finish directly
		if err := gsm.Transition(ctx, EventFinish); err != nil {
			t.Fatalf("Failed to finish: %v", err)
		}
		if gsm.GetCurrentState() != StateFinished {
			t.Errorf("Expected %s, got %s", StateFinished, gsm.GetCurrentState())
		}
	})

	// Path 3: Preparing -> Running -> Paused -> Finished
	t.Run("Path3_Preparing_Running_Paused_Finished", func(t *testing.T) {
		game := &model.Game{
			ID:           3,
			Status:       "draft",
			CurrentPhase: "preparation",
		}
		gsm := NewGameStateMachine(game, eb, false)

		// Start
		if err := gsm.Transition(ctx, EventStart); err != nil {
			t.Fatalf("Failed to start: %v", err)
		}

		// Pause
		if err := gsm.Transition(ctx, EventPause); err != nil {
			t.Fatalf("Failed to pause: %v", err)
		}

		// Finish from paused
		if err := gsm.Transition(ctx, EventFinish); err != nil {
			t.Fatalf("Failed to finish: %v", err)
		}
		if gsm.GetCurrentState() != StateFinished {
			t.Errorf("Expected %s, got %s", StateFinished, gsm.GetCurrentState())
		}
	})
}

// TestGameStateMachine_Callback_WithError tests callback error handling
func TestGameStateMachine_Callback_WithError(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	// Add callback that fails
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		return ErrTransitionCallback
	})

	// Try transition - should fail
	err := gsm.Transition(ctx, EventStart)
	if err == nil {
		t.Error("Expected error from callback, got nil")
	}

	// State should not change
	if gsm.GetCurrentState() != StatePreparing {
		t.Errorf("State should remain %s, got %s", StatePreparing, gsm.GetCurrentState())
	}
}

// TestGameStateMachine_ConcurrentSafety tests concurrent state transitions
func TestGameStateMachine_ConcurrentSafety(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	// Try concurrent transitions
	done := make(chan bool, 100)
	var successCount int
	var mu sync.Mutex

	for i := 0; i < 100; i++ {
		go func() {
			err := gsm.Transition(ctx, EventStart)
			if err == nil {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	// Only one should succeed
	if successCount != 1 {
		t.Errorf("Expected 1 successful transition, got %d", successCount)
	}

	// Final state should be running
	if gsm.GetCurrentState() != StateRunning {
		t.Errorf("Final state should be %s, got %s", StateRunning, gsm.GetCurrentState())
	}
}

// TestGameStateMachine_StateRecovery tests state recovery after creation
func TestGameStateMachine_StateRecovery(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		currentPhase  string
		expectedState GameState
	}{
		{"Recover from running state", "active", "running", StateRunning},
		{"Recover from paused state", "active", "break", StatePaused},
		{"Recover from finished state", "finished", "finished", StateFinished},
		{"Recover from preparing state", "draft", "preparation", StatePreparing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := &model.Game{
				ID:           1,
				Status:       tt.status,
				CurrentPhase: tt.currentPhase,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Create state machine (simulates restart)
			gsm := NewGameStateMachine(game, nil, false)

			// Verify state was recovered
			if gsm.GetCurrentState() != tt.expectedState {
				t.Errorf("Expected state %s after recovery, got %s", tt.expectedState, gsm.GetCurrentState())
			}

			// Verify game model is correct
			if gsm.GetGame().Status != tt.status {
				t.Errorf("Expected status %s, got %s", tt.status, gsm.GetGame().Status)
			}
			if gsm.GetGame().CurrentPhase != tt.currentPhase {
				t.Errorf("Expected phase %s, got %s", tt.currentPhase, gsm.GetGame().CurrentPhase)
			}
		})
	}
}

// TestGameStateMachine_MultipleCallbacks_AllPass tests all callbacks pass
func TestGameStateMachine_MultipleCallbacks_AllPass(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	var callCount int

	// Add 5 callbacks
	for i := 0; i < 5; i++ {
		gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
			callCount++
			return nil
		})
	}

	// Transition
	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// All callbacks should be called
	if callCount != 5 {
		t.Errorf("Expected 5 callback calls, got %d", callCount)
	}
}

// TestGameStateMachine_CallbackContext tests context is passed correctly
func TestGameStateMachine_CallbackContext(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	var receivedCtx context.Context

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		receivedCtx = ctx
		return nil
	})

	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if receivedCtx != ctx {
		t.Error("Context not passed correctly to callback")
	}
}

// TestGameStateMachine_CallbackGameID tests game ID is passed correctly
func TestGameStateMachine_CallbackGameID(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           42,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	var receivedGameID int64

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		receivedGameID = gameID
		return nil
	})

	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if receivedGameID != 42 {
		t.Errorf("Expected game ID 42, got %d", receivedGameID)
	}
}

// TestGameStateMachine_TransitionEvent tests event is passed correctly
func TestGameStateMachine_TransitionEvent(t *testing.T) {
	ctx := context.Background()
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	var receivedEvent GameEvent

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		receivedEvent = event
		return nil
	})

	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if receivedEvent != EventStart {
		t.Errorf("Expected event %s, got %s", EventStart, receivedEvent)
	}
}

// TestGameStateMachine_StateConstants tests state string constants
func TestGameStateMachine_StateConstants(t *testing.T) {
	if StatePreparing != "preparing" {
		t.Errorf("StatePreparing should be 'preparing', got %s", StatePreparing)
	}
	if StateRunning != "running" {
		t.Errorf("StateRunning should be 'running', got %s", StateRunning)
	}
	if StatePaused != "paused" {
		t.Errorf("StatePaused should be 'paused', got %s", StatePaused)
	}
	if StateFinished != "finished" {
		t.Errorf("StateFinished should be 'finished', got %s", StateFinished)
	}
}

// TestGameStateMachine_EventConstants tests event string constants
func TestGameStateMachine_EventConstants(t *testing.T) {
	if EventStart != "start" {
		t.Errorf("EventStart should be 'start', got %s", EventStart)
	}
	if EventPause != "pause" {
		t.Errorf("EventPause should be 'pause', got %s", EventPause)
	}
	if EventResume != "resume" {
		t.Errorf("EventResume should be 'resume', got %s", EventResume)
	}
	if EventFinish != "finish" {
		t.Errorf("EventFinish should be 'finish', got %s", EventFinish)
	}
}

// TestGameStateMachine_ErrorConstants tests error constants
func TestGameStateMachine_ErrorConstants(t *testing.T) {
	if ErrInvalidTransition == nil {
		t.Error("ErrInvalidTransition should not be nil")
	}
	if ErrGameNotFound == nil {
		t.Error("ErrGameNotFound should not be nil")
	}
	if ErrInvalidCurrentState == nil {
		t.Error("ErrInvalidCurrentState should not be nil")
	}
	if ErrTransitionCallback == nil {
		t.Error("ErrTransitionCallback should not be nil")
	}
}

// TestGameStateMachine_GetValidTransitions_Immutability tests that GetValidTransitions returns a copy
func TestGameStateMachine_GetValidTransitions_Immutability(t *testing.T) {
	game := &model.Game{
		ID:           1,
		Status:       "draft",
		CurrentPhase: "preparation",
	}
	gsm := NewGameStateMachine(game, nil, false)

	// Get transitions
	transitions1 := gsm.GetValidTransitions()
	transitions2 := gsm.GetValidTransitions()

	// Modify first map
	transitions1[EventStart] = StateFinished // This shouldn't affect the state machine

	// Get transitions again
	transitions3 := gsm.GetValidTransitions()

	// Verify state machine's valid transitions are unchanged
	if transitions3[EventStart] != StateRunning {
		t.Error("GetValidTransitions should return a copy, not a reference")
	}

	// Verify each call returns a new map
	if &transitions1 == &transitions2 {
		t.Error("GetValidTransitions should return a new map each time")
	}
}
