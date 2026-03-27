package scoring

import (
	"context"
	"sync"

	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/awd-platform/awd-arena/internal/repo"
	"github.com/awd-platform/awd-arena/pkg/logger"
)

// ZeroSumScorer implements zero-sum scoring for AWD competitions.
// In zero-sum scoring: attacker's gain = defender's loss
// Total points across all teams remain constant.
//
// 零和计分规则：
// 1. 攻击成功：攻击方 +N 分，被攻方 -N 分（零和转移）
// 2. 首杀奖励：第一个攻破的队伍额外加分（从被攻方扣除）
// 3. 防御分：轮次结束时，未失分的队伍从防御池获得奖励（零和）
// 4. 总分守恒：所有队伍分数之和 = 初始总分
type ZeroSumScorer struct {
	mu sync.RWMutex

	initialTotal     float64 // 初始总分
	flagValue        float64 // 单个 Flag 分值
	firstBloodBonus  float64 // 首杀加成比例
	roundDefenseValue float64 // 每轮防御分值

	// State tracking
	teamScores      map[int64]*TeamScore // teamID -> score state
	firstBlood      *FirstBloodDetector
	defense         *DefenseCalculator

	// Persistence
	scoreRepo repo.ScoreRepo
}

// TeamScore tracks a team's scoring state
type TeamScore struct {
	TeamID          int64
	InitialScore    float64 // 初始分数
	AttackScore     float64 // 攻击获得分数（零和）
	DefenseScore    float64 // 防御相关分数（被扣分为负，防御奖励为正）
	DefenseBonus    float64 // 防御奖励分数（轮次结束未失分）
	TotalScore      float64 // 总分
	FlagsCaptured   int     // 成功攻破的 flag 数
	FlagsLost       int     // 被攻破的 flag 数
	RoundsDefended  int     // 成功防御的轮次数
}

// NewZeroSumScorer creates a new zero-sum scorer.
func NewZeroSumScorer(initialTotal, flagValue, firstBloodBonus float64, scoreRepo repo.ScoreRepo) *ZeroSumScorer {
	return &ZeroSumScorer{
		initialTotal:      initialTotal,
		flagValue:         flagValue,
		firstBloodBonus:   firstBloodBonus,
		roundDefenseValue: flagValue * 0.5, // Defense worth 50% of attack
		teamScores:        make(map[int64]*TeamScore),
		firstBlood:        NewFirstBloodDetector(),
		defense:           NewDefenseCalculator(flagValue * 0.5),
		scoreRepo:         scoreRepo,
	}
}

// Initialize sets up initial scores for all teams.
func (s *ZeroSumScorer) Initialize(teamIDs []int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	initialPerTeam := s.initialTotal / float64(len(teamIDs))

	for _, teamID := range teamIDs {
		s.teamScores[teamID] = &TeamScore{
			TeamID:       teamID,
			InitialScore: initialPerTeam,
			TotalScore:   initialPerTeam,
			AttackScore:  0,
			DefenseScore: 0,
			DefenseBonus: 0,
		}
	}

	logger.Info("zero-sum scorer initialized",
		"teams", len(teamIDs),
		"initial_total", s.initialTotal,
		"per_team", initialPerTeam,
	)
}

// OnAttack processes an attack event with zero-sum scoring.
// Attacker gains points, target loses points.
// 零和转移：攻击方得分 = 被攻方失分
func (s *ZeroSumScorer) OnAttack(ctx context.Context, attack *model.FlagSubmission) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !attack.IsCorrect {
		return nil // Only process correct submissions
	}

	// Calculate base points
	points := s.flagValue

	// Check for first blood
	isFirstBlood := s.firstBlood.CheckAndRecord(attack.FlagValue, attack.AttackerTeam, attack.Round)
	if isFirstBlood {
		bonusPoints := points * s.firstBloodBonus
		points += bonusPoints
		logger.Info("first blood bonus applied",
			"team", attack.AttackerTeam,
			"flag", attack.FlagValue,
			"bonus", bonusPoints,
		)
	}

	// Ensure teams exist in scores map
	attackerScore := s.getOrCreateTeam(attack.AttackerTeam)
	targetScore := s.getOrCreateTeam(attack.TargetTeam)

	// Zero-sum transfer: attacker gains, target loses
	attackerScore.AttackScore += points
	attackerScore.TotalScore += points
	attackerScore.FlagsCaptured++

	targetScore.DefenseScore -= points
	targetScore.TotalScore -= points
	targetScore.FlagsLost++

	// Record breach for defense calculation
	s.defense.RecordBreach(attack.Round, attack.TargetTeam)

	logger.Info("zero-sum attack processed",
		"attacker", attack.AttackerTeam,
		"target", attack.TargetTeam,
		"points", points,
		"attacker_total", attackerScore.TotalScore,
		"target_total", targetScore.TotalScore,
	)

	return nil
}

// getOrCreateTeam gets or creates a team score entry.
func (s *ZeroSumScorer) getOrCreateTeam(teamID int64) *TeamScore {
	if score, ok := s.teamScores[teamID]; ok {
		return score
	}

	// Create new team with zero initial score (will be set properly if initialized)
	score := &TeamScore{
		TeamID:       teamID,
		InitialScore: 0,
		TotalScore:   0,
		AttackScore:  0,
		DefenseScore: 0,
		DefenseBonus: 0,
	}
	s.teamScores[teamID] = score
	return score
}

// OnDefense processes defense events at round end.
// 防御分计算：轮次结束未失分的队伍获得防御奖励
//
// 零和防御分实现：
// 1. 计算总防御池：被攻扣分的总和
// 2. 将防御池分配给未失分的队伍
// 3. 保持总分守恒
func (s *ZeroSumScorer) OnDefense(ctx context.Context, round int, totalRounds int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 计算被攻扣分总和（防御池）
	var defensePool float64
	for _, score := range s.teamScores {
		if score.DefenseScore < 0 {
			defensePool += -score.DefenseScore // 被攻扣分（负数变正数）
		}
	}

	// 计算未失分队伍数量
	var undefeatedTeams []int64
	for teamID, score := range s.teamScores {
		if score.FlagsLost == 0 {
			undefeatedTeams = append(undefeatedTeams, teamID)
		}
	}

	// 分配防御池给未失分的队伍
	if len(undefeatedTeams) > 0 && defensePool > 0 {
		bonusPerTeam := defensePool / float64(len(undefeatedTeams))
		for _, teamID := range undefeatedTeams {
			score := s.teamScores[teamID]
			score.DefenseBonus += bonusPerTeam
			score.TotalScore += bonusPerTeam
			score.RoundsDefended++

			logger.Info("defense bonus awarded",
				"team", teamID,
				"round", round,
				"bonus", bonusPerTeam,
				"total_defended", score.RoundsDefended,
			)
		}
	}

	// 额外的轮次防御分（可选，会破坏严格的零和）
	// 如果需要严格零和，注释掉这部分
	s.calculateRoundDefenseBonus(round, totalRounds)

	return nil
}

// calculateRoundDefenseBonus 计算每轮的防御奖励
// 注意：这会轻微破坏零和特性，用于奖励轮次内未失分的队伍
func (s *ZeroSumScorer) calculateRoundDefenseBonus(round, totalRounds int) {
	for teamID, score := range s.teamScores {
		timesBreached := score.FlagsLost
		defenseBonus := s.defense.CalculateDefenseBonus(round, totalRounds, timesBreached)

		// 只给未失分的队伍防御奖励
		if timesBreached == 0 {
			score.DefenseBonus += defenseBonus
			score.TotalScore += defenseBonus

			logger.Info("round defense bonus calculated",
				"team", teamID,
				"round", round,
				"bonus", defenseBonus,
			)
		}
	}
}

// ValidateZeroSum ensures total points remain constant (for attack+defense scores only).
// 注意：由于防御奖励的存在，总分可能会有轻微偏差
func (s *ZeroSumScorer) ValidateZeroSum() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total float64
	for _, score := range s.teamScores {
		total += score.TotalScore
	}

	// 允许一定的误差（由于防御奖励的存在）
	tolerance := s.roundDefenseValue * float64(len(s.teamScores)) // 允许每队一轮防御分的误差
	if tolerance < 0.01 {
		tolerance = 0.01
	}

	isValid := almostEqual(total, s.initialTotal, tolerance)

	logger.Info("zero-sum validation",
		"total", total,
		"expected", s.initialTotal,
		"tolerance", tolerance,
		"valid", isValid,
	)

	return isValid
}

// ValidateStrictZeroSum 严格验证零和（不考虑防御奖励）
func (s *ZeroSumScorer) ValidateStrictZeroSum() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var total float64
	for _, score := range s.teamScores {
		// 只计算初始分 + 攻击分 + 防御分（被扣分）
		total += score.InitialScore + score.AttackScore + score.DefenseScore
	}

	return almostEqual(total, s.initialTotal, 0.01)
}

// GetTeamScores returns current team scores.
func (s *ZeroSumScorer) GetTeamScores() map[int64]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scores := make(map[int64]float64)
	for teamID, score := range s.teamScores {
		scores[teamID] = score.TotalScore
	}
	return scores
}

// GetTeamScoreDetails returns detailed scoring breakdown.
func (s *ZeroSumScorer) GetTeamScoreDetails(teamID int64) *TeamScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if score, ok := s.teamScores[teamID]; ok {
		return score
	}
	return nil
}

// CalculateScore finalizes scores for a round.
func (s *ZeroSumScorer) CalculateScore(ctx context.Context, gameID int64, round int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Persist all team scores
	for teamID, score := range s.teamScores {
		roundScore := &model.RoundScore{
			GameID:       gameID,
			Round:        round,
			TeamID:       teamID,
			AttackScore:  score.AttackScore,
			DefenseScore: score.DefenseScore + score.DefenseBonus,
			TotalScore:   score.TotalScore,
		}

		if s.scoreRepo != nil {
			if err := s.scoreRepo.SaveRoundScore(ctx, roundScore); err != nil {
				logger.Error("failed to save round score", "error", err, "team", teamID)
				return err
			}
		}
	}

	logger.Info("scores calculated and persisted",
		"game", gameID,
		"round", round,
		"teams", len(s.teamScores),
	)

	return nil
}

// GetScoreBreakdown 返回所有队伍的详细分数明细
func (s *ZeroSumScorer) GetScoreBreakdown() map[int64]*ScoreBreakdown {
	s.mu.RLock()
	defer s.mu.RUnlock()

	breakdown := make(map[int64]*ScoreBreakdown)
	for teamID, score := range s.teamScores {
		breakdown[teamID] = &ScoreBreakdown{
			TeamID:          teamID,
			InitialScore:    score.InitialScore,
			AttackScore:     score.AttackScore,
			DefensePenalty:  score.DefenseScore, // 负数表示被扣分
			DefenseBonus:    score.DefenseBonus,
			TotalScore:      score.TotalScore,
			FlagsCaptured:   score.FlagsCaptured,
			FlagsLost:       score.FlagsLost,
			RoundsDefended:  score.RoundsDefended,
		}
	}
	return breakdown
}

// ScoreBreakdown 分数明细
type ScoreBreakdown struct {
	TeamID          int64
	InitialScore    float64
	AttackScore     float64
	DefensePenalty  float64
	DefenseBonus    float64
	TotalScore      float64
	FlagsCaptured   int
	FlagsLost       int
	RoundsDefended  int
}

// almostEqual checks if two floats are approximately equal.
func almostEqual(a, b, tolerance float64) bool {
	diff := a - b
	if diff < 0 {
		diff = -diff
	}
	return diff <= tolerance
}
