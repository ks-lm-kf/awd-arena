package scoring

import (
	"context"
	"sync"
	"testing"

	"github.com/awd-platform/awd-arena/internal/model"
)

// MockScoreRepo is a mock implementation for testing.
type MockScoreRepo struct {
	SavedScores []*model.RoundScore
	mu          sync.Mutex
}

func (m *MockScoreRepo) SaveRoundScore(ctx context.Context, score *model.RoundScore) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.SavedScores = append(m.SavedScores, score)
	return nil
}

func (m *MockScoreRepo) GetRoundScores(ctx context.Context, gameID int64, round int) ([]*model.RoundScore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.SavedScores, nil
}

func (m *MockScoreRepo) GetTeamScores(ctx context.Context, gameID int64) ([]*model.RoundScore, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.SavedScores, nil
}

func TestZeroSumScorer_Initialize(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3, 4}
	scorer.Initialize(teamIDs)

	scores := scorer.GetTeamScores()
	if len(scores) != 4 {
		t.Errorf("Expected 4 teams, got %d", len(scores))
	}

	// Each team should start with 2500 points (10000 / 4)
	for teamID, score := range scores {
		if score != 2500 {
			t.Errorf("Team %d: expected initial score 2500, got %f", teamID, score)
		}
	}

	// Validate zero-sum
	if !scorer.ValidateZeroSum() {
		t.Error("Zero-sum validation failed after initialization")
	}

	// Validate strict zero-sum
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed after initialization")
	}
}

func TestZeroSumScorer_OnAttack(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Team 1 attacks Team 2's flag (first blood applies)
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}

	err := scorer.OnAttack(context.Background(), attack)
	if err != nil {
		t.Errorf("OnAttack failed: %v", err)
	}

	// Check scores
	details1 := scorer.GetTeamScoreDetails(1)
	details2 := scorer.GetTeamScoreDetails(2)

	// Team 1 should have gained 110 points (100 + 10% first blood bonus)
	if details1.AttackScore != 110 {
		t.Errorf("Team 1 attack score: expected 110 (with first blood), got %f", details1.AttackScore)
	}

	// Team 2 should have lost 110 points (zero-sum transfer)
	if details2.DefenseScore != -110 {
		t.Errorf("Team 2 defense score: expected -110, got %f", details2.DefenseScore)
	}

	// Validate strict zero-sum (attack + defense only)
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed after attack")
	}
}

func TestZeroSumScorer_OnAttackNoFirstBlood(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3}
	scorer.Initialize(teamIDs)

	// First attack (first blood)
	attack1 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack1)

	// Second attack on same flag by another team (no first blood)
	attack2 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 3,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack2)

	// Team 3 should have gained only 100 points (no first blood bonus)
	details3 := scorer.GetTeamScoreDetails(3)
	if details3.AttackScore != 100 {
		t.Errorf("Team 3 attack score: expected 100 (no first blood), got %f", details3.AttackScore)
	}

	// Validate strict zero-sum
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed after attacks")
	}
}

func TestZeroSumScorer_FirstBlood(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// First attack (first blood)
	attack1 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack1)

	// Team 1 should get 100 + 10% bonus = 110
	details1 := scorer.GetTeamScoreDetails(1)
	if details1.AttackScore != 110 {
		t.Errorf("First blood bonus: expected 110, got %f", details1.AttackScore)
	}

	// Second attack on same flag by different team (no first blood)
	attack2 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 2,
		TargetTeam:   1,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack2)

	// Team 2 should only get 100 (no bonus)
	details2 := scorer.GetTeamScoreDetails(2)
	if details2.AttackScore != 100 {
		t.Errorf("No first blood bonus: expected 100, got %f", details2.AttackScore)
	}
}

func TestZeroSumScorer_WrongFlag(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Wrong flag submission
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{wrong}",
		IsCorrect:    false,
	}

	err := scorer.OnAttack(context.Background(), attack)
	if err != nil {
		t.Errorf("OnAttack failed: %v", err)
	}

	// Scores should remain unchanged
	if !scorer.ValidateZeroSum() {
		t.Error("Zero-sum validation failed after wrong flag")
	}

	details1 := scorer.GetTeamScoreDetails(1)
	if details1.AttackScore != 0 {
		t.Errorf("Wrong flag should not affect score, got %f", details1.AttackScore)
	}
}

func TestZeroSumScorer_MultipleAttacks(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3}
	scorer.Initialize(teamIDs)

	// Multiple attacks
	for i := 0; i < 3; i++ {
		attack := &model.FlagSubmission{
			GameID:       1,
			Round:        1,
			AttackerTeam: 1,
			TargetTeam:   2,
			FlagValue:    "flag{test" + string(rune('A'+i)) + "}",
			IsCorrect:    true,
		}
		scorer.OnAttack(context.Background(), attack)
	}

	// Team 1 should have 3 × 100 = 300 points (one is first blood)
	details1 := scorer.GetTeamScoreDetails(1)
	if details1.FlagsCaptured != 3 {
		t.Errorf("Expected 3 flags captured, got %d", details1.FlagsCaptured)
	}

	// Validate strict zero-sum
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed after multiple attacks")
	}
}

// ==================== 防御分测试 ====================

func TestZeroSumScorer_OnDefense(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3}
	scorer.Initialize(teamIDs)

	// Team 1 attacks Team 2
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack)

	// Process defense at round end
	err := scorer.OnDefense(context.Background(), 1, 10)
	if err != nil {
		t.Errorf("OnDefense failed: %v", err)
	}

	// Team 3 (undefeated) should get defense bonus
	details3 := scorer.GetTeamScoreDetails(3)
	if details3.DefenseBonus <= 0 {
		t.Errorf("Team 3 should get defense bonus for not being attacked, got %f", details3.DefenseBonus)
	}
	if details3.RoundsDefended != 1 {
		t.Errorf("Team 3 should have 1 round defended, got %d", details3.RoundsDefended)
	}

	// Team 2 (defeated) should not get defense bonus from pool
	details2 := scorer.GetTeamScoreDetails(2)
	if details2.RoundsDefended != 0 {
		t.Errorf("Team 2 should have 0 rounds defended (was attacked), got %d", details2.RoundsDefended)
	}
}

func TestZeroSumScorer_DefensePool(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3, 4}
	scorer.Initialize(teamIDs)

	// Team 1 attacks Team 2 (110 points with first blood)
	attack1 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test123}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack1)

	// Team 1 attacks Team 3 (100 points, no first blood)
	attack2 := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   3,
		FlagValue:    "flag{test456}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack2)

	// Process defense - only Team 4 is undefeated
	scorer.OnDefense(context.Background(), 1, 10)

	// Team 4 should get all the defense pool
	details4 := scorer.GetTeamScoreDetails(4)
	if details4.DefenseBonus <= 0 {
		t.Errorf("Team 4 should get defense bonus, got %f", details4.DefenseBonus)
	}

	// Check that Team 2 and 3 (attacked) don't have defense rounds
	details2 := scorer.GetTeamScoreDetails(2)
	details3 := scorer.GetTeamScoreDetails(3)
	if details2.RoundsDefended != 0 || details3.RoundsDefended != 0 {
		t.Error("Attacked teams should not have defense rounds counted")
	}
}

// ==================== 并发安全测试 ====================

func TestZeroSumScorer_ConcurrentAttacks(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2, 3, 4}
	scorer.Initialize(teamIDs)

	var wg sync.WaitGroup
	errors := make(chan error, 100)

	// Concurrent attacks
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			attack := &model.FlagSubmission{
				GameID:       1,
				Round:        1,
				AttackerTeam: int64(idx%4 + 1),
				TargetTeam:   int64((idx+1)%4 + 1),
				FlagValue:    "flag{test" + string(rune('A'+idx%26)) + "}",
				IsCorrect:    true,
			}
			if err := scorer.OnAttack(context.Background(), attack); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Errorf("Concurrent attack error: %v", err)
	}

	// Validate zero-sum after concurrent attacks
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed after concurrent attacks")
	}
}

// ==================== 边界情况测试 ====================

func TestZeroSumScorer_EmptyTeams(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	// Initialize with empty teams (edge case)
	scorer.Initialize([]int64{})

	// Should handle gracefully
	if len(scorer.GetTeamScores()) != 0 {
		t.Error("Should have no teams after empty initialization")
	}
}

func TestZeroSumScorer_SingleTeam(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1}
	scorer.Initialize(teamIDs)

	scores := scorer.GetTeamScores()
	if len(scores) != 1 {
		t.Errorf("Expected 1 team, got %d", len(scores))
	}

	// Single team should have all points
	if scores[1] != 10000 {
		t.Errorf("Single team should have 10000 points, got %f", scores[1])
	}
}

func TestZeroSumScorer_NegativeScore(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(100, 100, 0.1, repo) // Very small initial pool

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Team 1 attacks Team 2 multiple times
	for i := 0; i < 10; i++ {
		attack := &model.FlagSubmission{
			GameID:       1,
			Round:        1,
			AttackerTeam: 1,
			TargetTeam:   2,
			FlagValue:    "flag{test" + string(rune('A'+i)) + "}",
			IsCorrect:    true,
		}
		scorer.OnAttack(context.Background(), attack)
	}

	// Team 2 should have negative score
	details2 := scorer.GetTeamScoreDetails(2)
	if details2.TotalScore >= 0 {
		t.Errorf("Team 2 should have negative score, got %f", details2.TotalScore)
	}

	// Validate strict zero-sum
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation should still hold with negative scores")
	}
}

func TestZeroSumScorer_LargeTeamCount(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(100000, 100, 0.1, repo)

	// 100 teams
	teamIDs := make([]int64, 100)
	for i := range teamIDs {
		teamIDs[i] = int64(i + 1)
	}
	scorer.Initialize(teamIDs)

	scores := scorer.GetTeamScores()
	if len(scores) != 100 {
		t.Errorf("Expected 100 teams, got %d", len(scores))
	}

	// Each team should have 1000 points
	for teamID, score := range scores {
		if score != 1000 {
			t.Errorf("Team %d: expected 1000 points, got %f", teamID, score)
		}
	}

	// Validate zero-sum
	if !scorer.ValidateZeroSum() {
		t.Error("Zero-sum validation failed with large team count")
	}
}

func TestZeroSumScorer_ScoreBreakdown(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Attack
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack)

	// Get breakdown
	breakdown := scorer.GetScoreBreakdown()
	if len(breakdown) != 2 {
		t.Errorf("Expected 2 breakdowns, got %d", len(breakdown))
	}

	team1Breakdown := breakdown[1]
	if team1Breakdown.AttackScore != 110 {
		t.Errorf("Team 1 attack score in breakdown: expected 110, got %f", team1Breakdown.AttackScore)
	}
	if team1Breakdown.FlagsCaptured != 1 {
		t.Errorf("Team 1 should have 1 flag captured, got %d", team1Breakdown.FlagsCaptured)
	}

	team2Breakdown := breakdown[2]
	if team2Breakdown.FlagsLost != 1 {
		t.Errorf("Team 2 should have 1 flag lost, got %d", team2Breakdown.FlagsLost)
	}
}

func TestZeroSumScorer_CalculateScore(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.1, repo)

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Attack
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack)

	// Calculate and persist scores
	err := scorer.CalculateScore(context.Background(), 1, 1)
	if err != nil {
		t.Errorf("CalculateScore failed: %v", err)
	}

	// Check saved scores
	if len(repo.SavedScores) != 2 {
		t.Errorf("Expected 2 saved scores, got %d", len(repo.SavedScores))
	}

	for _, saved := range repo.SavedScores {
		if saved.Round != 1 {
			t.Errorf("Saved score should be for round 1, got %d", saved.Round)
		}
		if saved.GameID != 1 {
			t.Errorf("Saved score should be for game 1, got %d", saved.GameID)
		}
	}
}

func TestZeroSumScorer_ZeroFirstBloodBonus(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0, repo) // No first blood bonus

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Attack (would be first blood, but no bonus)
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack)

	// Team 1 should have exactly 100 points (no bonus)
	details1 := scorer.GetTeamScoreDetails(1)
	if details1.AttackScore != 100 {
		t.Errorf("With zero bonus: expected 100, got %f", details1.AttackScore)
	}
}

func TestZeroSumScorer_HighFirstBloodBonus(t *testing.T) {
	repo := &MockScoreRepo{}
	scorer := NewZeroSumScorer(10000, 100, 0.5, repo) // 50% first blood bonus

	teamIDs := []int64{1, 2}
	scorer.Initialize(teamIDs)

	// Attack with high first blood bonus
	attack := &model.FlagSubmission{
		GameID:       1,
		Round:        1,
		AttackerTeam: 1,
		TargetTeam:   2,
		FlagValue:    "flag{test}",
		IsCorrect:    true,
	}
	scorer.OnAttack(context.Background(), attack)

	// Team 1 should have 150 points (100 + 50%)
	details1 := scorer.GetTeamScoreDetails(1)
	if details1.AttackScore != 150 {
		t.Errorf("With 50%% bonus: expected 150, got %f", details1.AttackScore)
	}

	// Validate strict zero-sum
	if !scorer.ValidateStrictZeroSum() {
		t.Error("Strict zero-sum validation failed with high bonus")
	}
}
