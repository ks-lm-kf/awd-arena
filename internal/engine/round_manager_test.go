package engine

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/awd-platform/awd-arena/internal/model"
)

func TestNewRoundManager(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   3,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	
	if manager.GetCurrentRound() != 0 {
		t.Errorf("expected initial round 0, got %d", manager.GetCurrentRound())
	}
	
	if manager.GetPhase() != PhasePreparation {
		t.Errorf("expected phase preparation, got %s", manager.GetPhase())
	}
	
	if manager.IsRunning() {
		t.Error("expected manager to not be running initially")
	}
	
	if manager.IsPaused() {
		t.Error("expected manager to not be paused initially")
	}
}

func TestRoundManager_StartStop(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   3,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	ctx := context.Background()
	
	// Test start
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	
	if !manager.IsRunning() {
		t.Error("expected manager to be running after start")
	}
	
	if manager.GetCurrentRound() != 1 {
		t.Errorf("expected current round 1, got %d", manager.GetCurrentRound())
	}
	
	// Test double start (should be no-op)
	if err := manager.Start(ctx); err != nil {
		t.Errorf("double start should not error, got: %v", err)
	}
	
	// Test stop
	if err := manager.Stop(ctx); err != nil {
		t.Fatalf("failed to stop manager: %v", err)
	}
	
	if manager.IsRunning() {
		t.Error("expected manager to not be running after stop")
	}
}

func TestRoundManager_PauseResume(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   3,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	ctx := context.Background()
	
	// Start first
	if err := manager.Start(ctx); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}
	
	// Test pause
	if err := manager.Pause(); err != nil {
		t.Fatalf("failed to pause manager: %v", err)
	}
	
	if !manager.IsPaused() {
		t.Error("expected manager to be paused")
	}
	
	if manager.GetPhase() != PhasePaused {
		t.Errorf("expected phase paused, got %s", manager.GetPhase())
	}
	
	// Test double pause (should be no-op)
	if err := manager.Pause(); err != nil {
		t.Errorf("double pause should not error, got: %v", err)
	}
	
	// Test resume
	if err := manager.Resume(ctx); err != nil {
		t.Fatalf("failed to resume manager: %v", err)
	}
	
	if manager.IsPaused() {
		t.Error("expected manager to not be paused after resume")
	}
	
	// Clean up
	manager.Stop(ctx)
}

func TestRoundManager_GetState(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   3,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	// Get initial state
	state := manager.GetState()
	
	if state.CurrentRound != 0 {
		t.Errorf("expected initial round 0, got %d", state.CurrentRound)
	}
	
	if state.Phase != PhasePreparation {
		t.Errorf("expected phase preparation, got %s", state.Phase)
	}
	
	if state.RoundDuration != 300 {
		t.Errorf("expected round duration 300, got %d", state.RoundDuration)
	}
	
	if state.BreakDuration != 60 {
		t.Errorf("expected break duration 60, got %d", state.BreakDuration)
	}
	
	// Start and check state
	ctx := context.Background()
	manager.Start(ctx)
	
	time.Sleep(100 * time.Millisecond) // Let it settle
	
	state = manager.GetState()
	
	if state.CurrentRound != 1 {
		t.Errorf("expected current round 1, got %d", state.CurrentRound)
	}
	
	if state.Phase != PhaseRunning {
		t.Errorf("expected phase running, got %s", state.Phase)
	}
	
	if state.RemainingTime <= 0 || state.RemainingTime > 300 {
		t.Errorf("expected remaining time between 1 and 300, got %d", state.RemainingTime)
	}
	
	manager.Stop(ctx)
}

func TestRoundManager_UpdateDurations(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   3,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	// Update durations
	manager.UpdateDurations(600, 120)
	
	state := manager.GetState()
	
	if state.RoundDuration != 600 {
		t.Errorf("expected round duration 600, got %d", state.RoundDuration)
	}
	
	if state.BreakDuration != 120 {
		t.Errorf("expected break duration 120, got %d", state.BreakDuration)
	}
}

func TestRoundManager_Callbacks(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 1, // 1 second for fast testing
		BreakDuration: 1,
		TotalRounds:   2,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	var mu sync.Mutex
	roundStarts := make([]int, 0)
	roundEnds := make([]int, 0)
	
	// Set callbacks
	manager.SetCallbacks(
		func(ctx context.Context, round int) error {
			mu.Lock()
			roundStarts = append(roundStarts, round)
			mu.Unlock()
			return nil
		},
		func(ctx context.Context, round int) error {
			mu.Lock()
			roundEnds = append(roundEnds, round)
			mu.Unlock()
			return nil
		},
	)
	
	ctx := context.Background()
	manager.Start(ctx)
	
	// Wait for round 1 to complete (1s round + 1s break = 2s, but add buffer)
	time.Sleep(3 * time.Second)
	
	manager.Stop(ctx)
	
	// Verify callbacks were called
	mu.Lock()
	defer mu.Unlock()
	
	if len(roundStarts) < 1 {
		t.Error("expected at least one round start callback")
	} else if roundStarts[0] != 1 {
		t.Errorf("expected first round start to be round 1, got %d", roundStarts[0])
	}
	
	if len(roundEnds) < 1 {
		t.Error("expected at least one round end callback")
	} else if roundEnds[0] != 1 {
		t.Errorf("expected first round end to be round 1, got %d", roundEnds[0])
	}
}

func TestRoundManager_PauseAdjustsTiming(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 10,
		BreakDuration: 5,
		TotalRounds:   1,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	ctx := context.Background()
	manager.Start(ctx)
	
	// Let some time pass
	time.Sleep(2 * time.Second)
	
	state1 := manager.GetState()
	
	// Pause for a while
	manager.Pause()
	time.Sleep(2 * time.Second)
	
	// Resume
	manager.Resume(ctx)
	
	state2 := manager.GetState()
	
	// Remaining time should be approximately the same (within 1 second tolerance)
	// because we paused and resumed
	diff := state1.RemainingTime - state2.RemainingTime
	if diff > 1 || diff < -1 {
		t.Errorf("remaining time changed unexpectedly: before=%d, after=%d, diff=%d",
			state1.RemainingTime, state2.RemainingTime, diff)
	}
	
	manager.Stop(ctx)
}

func TestRoundManager_ConcurrentAccess(t *testing.T) {
	game := &model.Game{
		ID:            1,
		Title:         "Test Game",
		RoundDuration: 300,
		BreakDuration: 60,
		TotalRounds:   10,
	}
	
	engine, _ := NewCompetitionEngine(game)
	manager := NewRoundManager(game, engine)
	
	ctx := context.Background()
	
	// Start manager
	manager.Start(ctx)
	
	// Concurrent reads and writes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(3)
		
		// Reader 1
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = manager.GetState()
				_ = manager.GetCurrentRound()
				_ = manager.GetPhase()
			}
		}()
		
		// Reader 2
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = manager.IsRunning()
				_ = manager.IsPaused()
			}
		}()
		
		// Writer
		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				manager.Pause()
				time.Sleep(time.Millisecond)
				manager.Resume(ctx)
				time.Sleep(time.Millisecond)
			}
		}()
	}
	
	wg.Wait()
	manager.Stop(ctx)
}

func TestRoundPhase_String(t *testing.T) {
	phases := []RoundPhase{
		PhasePreparation,
		PhaseRunning,
		PhaseBreak,
		PhasePaused,
		PhaseFinished,
		PhaseScoring,
	}
	
	for _, phase := range phases {
		if string(phase) == "" {
			t.Errorf("phase %v should have non-empty string representation", phase)
		}
	}
}
