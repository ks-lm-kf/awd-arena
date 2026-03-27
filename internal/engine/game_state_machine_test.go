package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
)

// mockDatabase is a simple in-memory database for testing
var mockGames = make(map[int64]*model.Game)
var mockGameIDCounter int64 = 1

func setupMockGame(status, currentPhase string) *model.Game {
	game := &model.Game{
		ID:           mockGameIDCounter,
		Title:        "Test Game",
		Status:       status,
		CurrentPhase: currentPhase,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	mockGames[game.ID] = game
	mockGameIDCounter++
	return game
}

func resetMockDB() {
	mockGames = make(map[int64]*model.Game)
	mockGameIDCounter = 1
}

func TestNewGameStateMachine_NilGame(t *testing.T) {
	gsm := NewGameStateMachine(nil, nil, false)
	if gsm != nil {
		t.Error("Expected nil GameStateMachine for nil game")
	}
}

func TestNewGameStateMachine_InitialState(t *testing.T) {
	resetMockDB()
	tests := []struct {
		name          string
		status        string
		currentPhase  string
		expectedState GameState
	}{
		{"Draft/Preparation", "draft", "preparation", StatePreparing},
		{"Active/Running", "active", "running", StateRunning},
		{"Active/Break", "active", "break", StatePaused},
		{"Finished", "finished", "finished", StateFinished},
		{"Unknown status defaults to preparing", "unknown", "", StatePreparing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := setupMockGame(tt.status, tt.currentPhase)
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

func TestGameStateMachine_Transition_ValidTransitions(t *testing.T) {
	resetMockDB()
	ctx := context.Background()

	tests := []struct {
		name          string
		initialState  GameState
		event         GameEvent
		expectedState GameState
	}{
		{"Preparing to Running", StatePreparing, EventStart, StateRunning},
		{"Running to Paused", StateRunning, EventPause, StatePaused},
		{"Paused to Running", StatePaused, EventResume, StateRunning},
		{"Running to Finished", StateRunning, EventFinish, StateFinished},
		{"Paused to Finished", StatePaused, EventFinish, StateFinished},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game with appropriate initial state
			var game *model.Game
			switch tt.initialState {
			case StatePreparing:
				game = setupMockGame("draft", "preparation")
			case StateRunning:
				game = setupMockGame("active", "running")
			case StatePaused:
				game = setupMockGame("active", "break")
			case StateFinished:
				game = setupMockGame("finished", "finished")
			}

			gsm := NewGameStateMachine(game, nil, false)

			err := gsm.Transition(ctx, tt.event)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if gsm.GetCurrentState() != tt.expectedState {
				t.Errorf("Expected state %s, got %s", tt.expectedState, gsm.GetCurrentState())
			}
		})
	}
}

func TestGameStateMachine_Transition_InvalidTransitions(t *testing.T) {
	resetMockDB()
	ctx := context.Background()

	tests := []struct {
		name         string
		initialState GameState
		event        GameEvent
	}{
		{"Cannot pause from preparing", StatePreparing, EventPause},
		{"Cannot resume from preparing", StatePreparing, EventResume},
		{"Cannot finish from preparing", StatePreparing, EventFinish},
		{"Cannot start from running", StateRunning, EventStart},
		{"Cannot resume from running", StateRunning, EventResume},
		{"Cannot start from paused", StatePaused, EventStart},
		{"Cannot pause from paused", StatePaused, EventPause},
		{"Cannot start from finished", StateFinished, EventStart},
		{"Cannot pause from finished", StateFinished, EventPause},
		{"Cannot resume from finished", StateFinished, EventResume},
		{"Cannot do anything from finished", StateFinished, EventFinish},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create game with appropriate initial state
			var game *model.Game
			switch tt.initialState {
			case StatePreparing:
				game = setupMockGame("draft", "preparation")
			case StateRunning:
				game = setupMockGame("active", "running")
			case StatePaused:
				game = setupMockGame("active", "break")
			case StateFinished:
				game = setupMockGame("finished", "finished")
			}

			gsm := NewGameStateMachine(game, nil, false)

			err := gsm.Transition(ctx, tt.event)
			if err == nil {
				t.Error("Expected error for invalid transition, got nil")
			}
			if gsm.GetCurrentState() != tt.initialState {
				t.Errorf("State should remain %s, got %s", tt.initialState, gsm.GetCurrentState())
			}
		})
	}
}

func TestGameStateMachine_Callbacks(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	var callbackCalled bool
	var callbackFromState, callbackToState GameState
	var callbackEvent GameEvent

	// Add callback
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callbackCalled = true
		callbackFromState = fromState
		callbackToState = toState
		callbackEvent = event
		return nil
	})

	// Trigger transition
	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify callback was called
	if !callbackCalled {
		t.Error("Callback was not called")
	}
	if callbackFromState != StatePreparing {
		t.Errorf("Expected from state %s, got %s", StatePreparing, callbackFromState)
	}
	if callbackToState != StateRunning {
		t.Errorf("Expected to state %s, got %s", StateRunning, callbackToState)
	}
	if callbackEvent != EventStart {
		t.Errorf("Expected event %s, got %s", EventStart, callbackEvent)
	}
}

func TestGameStateMachine_CallbackError(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	// Add failing callback
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		return errors.New("callback error")
	})

	// Try transition
	err := gsm.Transition(ctx, EventStart)
	if err == nil {
		t.Error("Expected error from callback, got nil")
	}

	// State should not change
	if gsm.GetCurrentState() != StatePreparing {
		t.Errorf("State should remain %s, got %s", StatePreparing, gsm.GetCurrentState())
	}
}

func TestGameStateMachine_MultipleCallbacks(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	var callOrder []int

	// Add multiple callbacks
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 1)
		return nil
	})

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 2)
		return nil
	})

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 3)
		return nil
	})

	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify callbacks were called in order
	if len(callOrder) != 3 {
		t.Errorf("Expected 3 callbacks, got %d", len(callOrder))
	}
	for i, expected := range []int{1, 2, 3} {
		if callOrder[i] != expected {
			t.Errorf("Callback %d: expected order %d, got %d", i, expected, callOrder[i])
		}
	}
}

func TestGameStateMachine_CallbackStopsOnSecondError(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	var callOrder []int

	// Add callbacks, second one fails
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 1)
		return nil
	})

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 2)
		return errors.New("callback 2 error")
	})

	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callOrder = append(callOrder, 3)
		return nil
	})

	err := gsm.Transition(ctx, EventStart)
	if err == nil {
		t.Error("Expected error from callback 2")
	}

	// First callback should be called, second should be called, third should not
	if len(callOrder) != 2 {
		t.Errorf("Expected 2 callbacks to be called, got %d", len(callOrder))
	}
	if callOrder[0] != 1 || callOrder[1] != 2 {
		t.Errorf("Expected callbacks [1, 2], got %v", callOrder)
	}
}

func TestGameStateMachine_GetValidTransitions(t *testing.T) {
	resetMockDB()
	tests := []struct {
		name               string
		initialState       GameState
		expectedCount      int
		expectedTransitions map[GameEvent]bool
	}{
		{"Preparing", StatePreparing, 1, map[GameEvent]bool{EventStart: true}},
		{"Running", StateRunning, 2, map[GameEvent]bool{EventPause: true, EventFinish: true}},
		{"Paused", StatePaused, 2, map[GameEvent]bool{EventResume: true, EventFinish: true}},
		{"Finished", StateFinished, 0, map[GameEvent]bool{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var game *model.Game
			switch tt.initialState {
			case StatePreparing:
				game = setupMockGame("draft", "preparation")
			case StateRunning:
				game = setupMockGame("active", "running")
			case StatePaused:
				game = setupMockGame("active", "break")
			case StateFinished:
				game = setupMockGame("finished", "finished")
			}

			gsm := NewGameStateMachine(game, nil, false)
			transitions := gsm.GetValidTransitions()

			if len(transitions) != tt.expectedCount {
				t.Errorf("Expected %d transitions, got %d", tt.expectedCount, len(transitions))
			}

			for event, shouldExist := range tt.expectedTransitions {
				_, exists := transitions[event]
				if shouldExist && !exists {
					t.Errorf("Expected transition %s to exist", event)
				}
				if !shouldExist && exists {
					t.Errorf("Expected transition %s to not exist", event)
				}
			}
		})
	}
}

func TestGameStateMachine_CanTransition(t *testing.T) {
	resetMockDB()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	// Can start from preparing
	if !gsm.CanTransition(EventStart) {
		t.Error("Should be able to start from preparing state")
	}

	// Cannot pause from preparing
	if gsm.CanTransition(EventPause) {
		t.Error("Should not be able to pause from preparing state")
	}

	// Cannot finish from preparing
	if gsm.CanTransition(EventFinish) {
		t.Error("Should not be able to finish from preparing state")
	}
}

func TestGameStateMachine_ConcurrentStateTransitions(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	// Try 10 concurrent transitions
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := gsm.Transition(ctx, EventStart)
			if err != nil {
				errors <- err
			}
		}()
	}

	wg.Wait()
	close(errors)

	// All but one should fail (mutex should prevent race)
	errorCount := 0
	for range errors {
		errorCount++
	}

	// Should have exactly 9 errors (first transition succeeds, rest fail)
	if errorCount != 9 {
		t.Errorf("Expected 9 errors from concurrent transitions, got %d", errorCount)
	}

	// Final state should be running
	if gsm.GetCurrentState() != StateRunning {
		t.Errorf("Final state should be %s, got %s", StateRunning, gsm.GetCurrentState())
	}
}

func TestGameStateMachine_EventBus(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")

	// Create event bus
	eb := &eventbus.Bus{}
	gsm := NewGameStateMachine(game, eb, false)

	// Note: The actual event bus in the project is a no-op implementation
	// This test verifies the state machine doesn't crash when an event bus is present

	// Trigger transition
	err := gsm.Transition(ctx, EventStart)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	// Verify state transition occurred
	if gsm.GetCurrentState() != StateRunning {
		t.Errorf("Expected state %s, got %s", StateRunning, gsm.GetCurrentState())
	}
}

func TestGameStateMachine_GetGame(t *testing.T) {
	resetMockDB()
	game := setupMockGame("draft", "preparation")
	gsm := NewGameStateMachine(game, nil, false)

	retrievedGame := gsm.GetGame()
	if retrievedGame == nil {
		t.Fatal("GetGame returned nil")
	}
	if retrievedGame.ID != game.ID {
		t.Errorf("Expected game ID %d, got %d", game.ID, retrievedGame.ID)
	}
}

func TestGameStateMachine_UpdateModelState(t *testing.T) {
	resetMockDB()
	tests := []struct {
		name               string
		state              GameState
		expectedStatus     string
		expectedPhase      string
	}{
		{"Preparing state", StatePreparing, "draft", "preparation"},
		{"Running state", StateRunning, "active", "running"},
		{"Paused state", StatePaused, "active", "break"},
		{"Finished state", StateFinished, "finished", "finished"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := setupMockGame("unknown", "unknown")
			gsm := NewGameStateMachine(game, nil, false)
			gsm.state = tt.state
			gsm.updateModelState()

			if gsm.game.Status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, gsm.game.Status)
			}
			if gsm.game.CurrentPhase != tt.expectedPhase {
				t.Errorf("Expected phase %s, got %s", tt.expectedPhase, gsm.game.CurrentPhase)
			}
		})
	}
}

func TestGetStateTransitions(t *testing.T) {
	transitions := GetStateTransitions()

	if transitions == nil {
		t.Fatal("GetStateTransitions returned nil")
	}

	// Verify structure
	if len(transitions[StatePreparing]) != 1 {
		t.Errorf("Preparing state should have 1 transition, got %d", len(transitions[StatePreparing]))
	}
	if len(transitions[StateRunning]) != 2 {
		t.Errorf("Running state should have 2 transitions, got %d", len(transitions[StateRunning]))
	}
	if len(transitions[StatePaused]) != 2 {
		t.Errorf("Paused state should have 2 transitions, got %d", len(transitions[StatePaused]))
	}
	if len(transitions[StateFinished]) != 0 {
		t.Errorf("Finished state should have 0 transitions, got %d", len(transitions[StateFinished]))
	}
}

func TestGameStateMachine_FullLifecycle(t *testing.T) {
	resetMockDB()
	ctx := context.Background()
	game := setupMockGame("draft", "preparation")
	eb := &eventbus.Bus{}
	gsm := NewGameStateMachine(game, eb, false)

	var callbackLog []string
	gsm.AddCallback(func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error {
		callbackLog = append(callbackLog, string(toState))
		return nil
	})

	// Start game
	if err := gsm.Transition(ctx, EventStart); err != nil {
		t.Fatalf("Failed to start game: %v", err)
	}
	if gsm.GetCurrentState() != StateRunning {
		t.Errorf("Expected state %s, got %s", StateRunning, gsm.GetCurrentState())
	}

	// Pause game
	if err := gsm.Transition(ctx, EventPause); err != nil {
		t.Fatalf("Failed to pause game: %v", err)
	}
	if gsm.GetCurrentState() != StatePaused {
		t.Errorf("Expected state %s, got %s", StatePaused, gsm.GetCurrentState())
	}

	// Resume game
	if err := gsm.Transition(ctx, EventResume); err != nil {
		t.Fatalf("Failed to resume game: %v", err)
	}
	if gsm.GetCurrentState() != StateRunning {
		t.Errorf("Expected state %s, got %s", StateRunning, gsm.GetCurrentState())
	}

	// Finish game
	if err := gsm.Transition(ctx, EventFinish); err != nil {
		t.Fatalf("Failed to finish game: %v", err)
	}
	if gsm.GetCurrentState() != StateFinished {
		t.Errorf("Expected state %s, got %s", StateFinished, gsm.GetCurrentState())
	}

	// Verify callback log
	expectedLog := []string{string(StateRunning), string(StatePaused), string(StateRunning), string(StateFinished)}
	if len(callbackLog) != len(expectedLog) {
		t.Errorf("Expected %d callbacks, got %d", len(expectedLog), len(callbackLog))
	}
	for i, expected := range expectedLog {
		if callbackLog[i] != expected {
			t.Errorf("Callback %d: expected %s, got %s", i, expected, callbackLog[i])
		}
	}
}
