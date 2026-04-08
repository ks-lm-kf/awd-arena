package engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// GameState represents the possible states of a game
type GameState string

const (
	// StatePreparing - Game is in preparation phase
	StatePreparing GameState = "preparing"
	// StateRunning - Game is in progress
	StateRunning GameState = "running"
	// StatePaused - Game is paused
	StatePaused GameState = "paused"
	// StateFinished - Game has ended
	StateFinished GameState = "finished"
)

// GameEvent represents events that trigger state transitions
type GameEvent string

const (
	// EventStart - Start the game
	EventStart GameEvent = "start"
	// EventPause - Pause the game
	EventPause GameEvent = "pause"
	// EventResume - Resume from paused state
	EventResume GameEvent = "resume"
	// EventFinish - End the game
	EventFinish GameEvent = "finish"
)

// StateTransitionCallback is called when a state transition occurs
type StateTransitionCallback func(ctx context.Context, gameID int64, fromState, toState GameState, event GameEvent) error

// GameStateMachine manages game state transitions
type GameStateMachine struct {
	mu        sync.RWMutex
	gameID    int64
	game      *model.Game
	state     GameState
	callbacks []StateTransitionCallback
	eventBus  *eventbus.Bus
	persistDB bool
}

// Valid state transitions
var validTransitions = map[GameState]map[GameEvent]GameState{
	StatePreparing: {
		EventStart: StateRunning,
	},
	StateRunning: {
		EventPause:  StatePaused,
		EventFinish: StateFinished,
	},
	StatePaused: {
		EventResume: StateRunning,
		EventFinish: StateFinished,
	},
	StateFinished: {}, // No transitions from finished state
}

// State transition errors
var (
	ErrInvalidTransition   = errors.New("invalid state transition")
	ErrGameNotFound        = errors.New("game not found")
	ErrInvalidCurrentState = errors.New("invalid current state")
	ErrTransitionCallback  = errors.New("transition callback failed")
)

// NewGameStateMachine creates a new game state machine
func NewGameStateMachine(game *model.Game, eventBus *eventbus.Bus, persistDB bool) *GameStateMachine {
	if game == nil {
		return nil
	}

	// Map model.Status and model.CurrentPhase to GameState
	initialState := mapModelToGameState(game.Status, game.CurrentPhase)

	return &GameStateMachine{
		gameID:    game.ID,
		game:      game,
		state:     initialState,
		callbacks: make([]StateTransitionCallback, 0),
		eventBus:  eventBus,
		persistDB: persistDB,
	}
}

// LoadGameStateMachine loads a game state machine from the database
func LoadGameStateMachine(ctx context.Context, gameID int64, eventBus *eventbus.Bus, persistDB bool) (*GameStateMachine, error) {
	db := database.GetDB()
	if db == nil {
		return nil, errors.New("database not initialized")
	}

	var game model.Game
	if err := db.WithContext(ctx).First(&game, gameID).Error; err != nil {
		return nil, ErrGameNotFound
	}

	return NewGameStateMachine(&game, eventBus, persistDB), nil
}

// AddCallback registers a callback for state transitions
func (gsm *GameStateMachine) AddCallback(callback StateTransitionCallback) {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()
	gsm.callbacks = append(gsm.callbacks, callback)
}

// GetCurrentState returns the current state
func (gsm *GameStateMachine) GetCurrentState() GameState {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.state
}

// GetGame returns the game model
func (gsm *GameStateMachine) GetGame() *model.Game {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()
	return gsm.game
}

// Transition attempts to transition to a new state based on the event
func (gsm *GameStateMachine) Transition(ctx context.Context, event GameEvent) error {
	gsm.mu.Lock()
	defer gsm.mu.Unlock()

	// Get allowed transitions for current state
	allowedTransitions, ok := validTransitions[gsm.state]
	if !ok {
		return ErrInvalidCurrentState
	}

	// Check if the event is valid for current state
	newState, ok := allowedTransitions[event]
	if !ok {
		return fmt.Errorf("%w: cannot %s from state %s", ErrInvalidTransition, event, gsm.state)
	}

	// Store old state for callbacks
	fromState := gsm.state

	// Execute pre-transition callbacks
	for i, callback := range gsm.callbacks {
		if err := callback(ctx, gsm.gameID, fromState, newState, event); err != nil {
			logger.Error("state transition callback failed",
				"callback_index", i,
				"from_state", fromState,
				"to_state", newState,
				"event", event,
				"error", err)
			return fmt.Errorf("%w: callback %d failed: %v", ErrTransitionCallback, i, err)
		}
	}

	// Update state
	gsm.state = newState

	// Update model
	gsm.updateModelState()

	// Persist to database if enabled
	if gsm.persistDB {
		if err := gsm.persistState(ctx); err != nil {
			logger.Error("failed to persist game state",
				"game_id", gsm.gameID,
				"state", newState,
				"error", err)
			// Rollback state
			gsm.state = fromState
			gsm.updateModelState()
			return fmt.Errorf("failed to persist state: %w", err)
		}
	}

	// Publish event to event bus
	if gsm.eventBus != nil {
		eventData := map[string]interface{}{
			"game_id":    gsm.gameID,
			"from_state": fromState,
			"to_state":   newState,
			"event":      event,
		}
		eventType := fmt.Sprintf("game_state_%s", newState)
		if err := gsm.eventBus.Publish(ctx, eventType, eventData); err != nil {
			logger.Error("failed to publish state change event",
				"game_id", gsm.gameID,
				"event_type", eventType,
				"error", err)
		}
	}

	logger.Info("game state transitioned",
		"game_id", gsm.gameID,
		"from_state", fromState,
		"to_state", newState,
		"event", event)

	return nil
}

// CanTransition checks if a transition is valid without executing it
func (gsm *GameStateMachine) CanTransition(event GameEvent) bool {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()

	allowedTransitions, ok := validTransitions[gsm.state]
	if !ok {
		return false
	}

	_, ok = allowedTransitions[event]
	return ok
}

// GetValidTransitions returns all valid transitions from current state
func (gsm *GameStateMachine) GetValidTransitions() map[GameEvent]GameState {
	gsm.mu.RLock()
	defer gsm.mu.RUnlock()

	result := make(map[GameEvent]GameState)
	if transitions, ok := validTransitions[gsm.state]; ok {
		for k, v := range transitions {
			result[k] = v
		}
	}
	return result
}

// updateModelState updates the game model based on state machine state
func (gsm *GameStateMachine) updateModelState() {
	switch gsm.state {
	case StatePreparing:
		gsm.game.Status = "draft"
		gsm.game.CurrentPhase = "preparation"
	case StateRunning:
		gsm.game.Status = "active"
		gsm.game.CurrentPhase = "running"
	case StatePaused:
		gsm.game.Status = "active"
		gsm.game.CurrentPhase = "break"
	case StateFinished:
		gsm.game.Status = "finished"
		gsm.game.CurrentPhase = "finished"
	}
}

// persistState saves the current state to the database
func (gsm *GameStateMachine) persistState(ctx context.Context) error {
	db := database.GetDB()
	if db == nil {
		return errors.New("database not initialized")
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        gsm.game.Status,
		"current_phase": gsm.game.CurrentPhase,
		"updated_at":    now,
	}

	// Set timestamps based on state
	switch gsm.state {
	case StateRunning:
		if gsm.game.StartTime == nil {
			updates["start_time"] = now
		}
	case StateFinished:
		updates["end_time"] = now
	}

	return db.WithContext(ctx).Model(&model.Game{}).
		Where("id = ?", gsm.gameID).
		Updates(updates).Error
}

// mapModelToGameState converts model fields to GameState
func mapModelToGameState(status, currentPhase string) GameState {
	switch status {
	case "finished":
		return StateFinished
	case "active":
		// Active games use currentPhase to distinguish running vs paused
		switch currentPhase {
		case "break":
			return StatePaused
		default:
			return StateRunning
		}
	case "draft":
		return StatePreparing
	default:
		return StatePreparing
	}
}

// GetStateHistory returns a string representation of valid state transitions
func GetStateTransitions() map[GameState]map[GameEvent]GameState {
	result := make(map[GameState]map[GameEvent]GameState)
	for k, v := range validTransitions {
		result[k] = make(map[GameEvent]GameState)
		for ek, ev := range v {
			result[k][ek] = ev
		}
	}
	return result
}
