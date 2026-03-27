package handler

import (
	"context"

	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
	"sort"
)

// LeaderboardHandler handles leaderboard-related requests
var LeaderboardHandler *leaderboardHandler

func init() {
	LeaderboardHandler = &leaderboardHandler{}
}

type leaderboardHandler struct{}

// RankHistoryEntry represents a single historical ranking entry
type RankHistoryEntry struct {
	Round      int64   `json:"round"`
	TeamID     int64   `json:"team_id"`
	TeamName   string  `json:"team_name"`
	Rank       int     `json:"rank"`
	TotalScore float64 `json:"total_score"`
}

// LeaderboardResponse represents the leaderboard response
type LeaderboardResponse struct {
	GameID  int64                  `json:"game_id"`
	Entries []model.LeaderboardEntry `json:"entries"`
}

// Get returns the current leaderboard for a game
// GET /api/v1/games/:id/leaderboard
func (h *leaderboardHandler) Get(c fiber.Ctx) error {
	gameID := parseID(c.Params("id"))

	entries, err := h.getLeaderboard(c, gameID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data": LeaderboardResponse{
			GameID:  gameID,
			Entries: entries,
		},
	})
}

// GetHistory returns the historical rankings for all rounds in a game
// GET /api/v1/games/:id/rankings/history
func (h *leaderboardHandler) GetHistory(c fiber.Ctx) error {
	gameID := parseID(c.Params("id"))

	entries, err := h.getRankHistory(c, gameID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"code":    500,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"code":    0,
		"message": "ok",
		"data":    entries,
	})
}

// getRankHistory returns historical rankings for all rounds
func (h *leaderboardHandler) getRankHistory(c fiber.Ctx, gameID int64) ([]RankHistoryEntry, error) {
	db := c.Locals("db").(*gorm.DB)

	// Get all teams
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Find(&gameTeams).Error; err != nil {
		return nil, err
	}

	teamIDs := make([]int64, len(gameTeams))
	for i, gt := range gameTeams {
		teamIDs[i] = gt.TeamID
	}

	var teams []model.Team
	if err := db.Where("id IN ?", teamIDs).Find(&teams).Error; err != nil {
		return nil, err
	}

	teamMap := make(map[int64]model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Get all round scores
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("round, team_id").Find(&roundScores).Error; err != nil {
		return nil, err
	}

	// Group by round
	roundMap := make(map[int][]model.RoundScore)
	for _, rs := range roundScores {
		roundMap[rs.Round] = append(roundMap[rs.Round], rs)
	}

	// Calculate rankings for each round
	var history []RankHistoryEntry
	for round := 1; round <= len(roundMap); round++ {
		scores, exists := roundMap[round]
		if !exists {
			continue
		}

		// Sort by total score
		sort.Slice(scores, func(i, j int) bool {
			return scores[i].TotalScore > scores[j].TotalScore
		})

		// Assign ranks
		for i, score := range scores {
			team, exists := teamMap[score.TeamID]
			teamName := ""
			if exists {
				teamName = team.Name
			}

			history = append(history, RankHistoryEntry{
				Round:      int64(round),
				TeamID:     score.TeamID,
				TeamName:   teamName,
				Rank:       i + 1,
				TotalScore: score.TotalScore,
			})
		}
	}

	return history, nil
}

// getLeaderboard calculates and returns the leaderboard entries
func (h *leaderboardHandler) getLeaderboard(c fiber.Ctx, gameID int64) ([]model.LeaderboardEntry, error) {
	db := c.Locals("db").(*gorm.DB)

	// Get all teams participating in the game
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Find(&gameTeams).Error; err != nil {
		return nil, err
	}

	// Get team details
	teamIDs := make([]int64, len(gameTeams))
	for i, gt := range gameTeams {
		teamIDs[i] = gt.TeamID
	}

	var teams []model.Team
	if err := db.Where("id IN ?", teamIDs).Find(&teams).Error; err != nil {
		return nil, err
	}

	teamMap := make(map[int64]model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Calculate scores from RoundScore
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Find(&roundScores).Error; err != nil {
		return nil, err
	}

	// Aggregate scores by team
	type teamStats struct {
		attackScore  float64
		defenseScore float64
		totalScore   float64
		firstBloods  int
	}
	teamScores := make(map[int64]*teamStats)

	for _, rs := range roundScores {
		if _, exists := teamScores[rs.TeamID]; !exists {
			teamScores[rs.TeamID] = &teamStats{}
		}
		teamScores[rs.TeamID].attackScore += rs.AttackScore
		teamScores[rs.TeamID].defenseScore += rs.DefenseScore
		teamScores[rs.TeamID].totalScore += rs.TotalScore
	}

	// Count first bloods from FlagSubmission
	type firstBloodCount struct {
		TeamID int64 `gorm:"column:attacker_team"`
		Count  int   `gorm:"column:count"`
	}
	var firstBloodCounts []firstBloodCount
	db.Table("flag_submissions").
		Select("attacker_team, COUNT(*) as count").
		Where("game_id = ? AND is_correct = ? AND is_first_blood = ?", gameID, true, true).
		Group("attacker_team").
		Scan(&firstBloodCounts)

	for _, fbc := range firstBloodCounts {
		if stats, exists := teamScores[fbc.TeamID]; exists {
			stats.firstBloods = fbc.Count
		}
	}

	// Build leaderboard entries
	var entries []model.LeaderboardEntry
	for teamID, stats := range teamScores {
		team, exists := teamMap[teamID]
		teamName := ""
		if exists {
			teamName = team.Name
		}
		entries = append(entries, model.LeaderboardEntry{
			TeamID:       teamID,
			TeamName:     teamName,
			TotalScore:   stats.totalScore,
			AttackScore:  stats.attackScore,
			DefenseScore: stats.defenseScore,
			FirstBloods:  stats.firstBloods,
		})
	}

	// Sort by total score descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalScore > entries[j].TotalScore
	})

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries, nil
}

// BroadcastUpdate broadcasts the current leaderboard to all WebSocket clients
func (h *leaderboardHandler) BroadcastUpdate(gameID int64, db *gorm.DB) error {
	// Create a fiber context alternative for getting leaderboard
	// We'll use a mock context or direct DB queries

	// Get all teams participating in the game
	var gameTeams []model.GameTeam
	if err := db.Where("game_id = ?", gameID).Find(&gameTeams).Error; err != nil {
		return err
	}

	// Get team details
	teamIDs := make([]int64, len(gameTeams))
	for i, gt := range gameTeams {
		teamIDs[i] = gt.TeamID
	}

	var teams []model.Team
	if err := db.Where("id IN ?", teamIDs).Find(&teams).Error; err != nil {
		return err
	}

	teamMap := make(map[int64]model.Team)
	for _, team := range teams {
		teamMap[team.ID] = team
	}

	// Calculate scores from RoundScore
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Find(&roundScores).Error; err != nil {
		return err
	}

	// Aggregate scores by team
	type teamStats struct {
		attackScore  float64
		defenseScore float64
		totalScore   float64
		firstBloods  int
	}
	teamScores := make(map[int64]*teamStats)

	for _, rs := range roundScores {
		if _, exists := teamScores[rs.TeamID]; !exists {
			teamScores[rs.TeamID] = &teamStats{}
		}
		teamScores[rs.TeamID].attackScore += rs.AttackScore
		teamScores[rs.TeamID].defenseScore += rs.DefenseScore
		teamScores[rs.TeamID].totalScore += rs.TotalScore
	}

	// Count first bloods
	type firstBloodCount struct {
		TeamID int64 `gorm:"column:attacker_team"`
		Count  int   `gorm:"column:count"`
	}
	var firstBloodCounts []firstBloodCount
	db.Table("flag_submissions").
		Select("attacker_team, COUNT(*) as count").
		Where("game_id = ? AND is_correct = ? AND is_first_blood = ?", gameID, true, true).
		Group("attacker_team").
		Scan(&firstBloodCounts)

	for _, fbc := range firstBloodCounts {
		if stats, exists := teamScores[fbc.TeamID]; exists {
			stats.firstBloods = fbc.Count
		}
	}

	// Build leaderboard entries
	var entries []model.LeaderboardEntry
	for teamID, stats := range teamScores {
		team, exists := teamMap[teamID]
		teamName := ""
		if exists {
			teamName = team.Name
		}
		entries = append(entries, model.LeaderboardEntry{
			TeamID:       teamID,
			TeamName:     teamName,
			TotalScore:   stats.totalScore,
			AttackScore:  stats.attackScore,
			DefenseScore: stats.defenseScore,
			FirstBloods:  stats.firstBloods,
		})
	}

	// Sort by total score descending
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].TotalScore > entries[j].TotalScore
	})

	// Assign ranks
	for i := range entries {
		entries[i].Rank = i + 1
	}

	// Broadcast to WebSocket clients via eventbus
	bus := eventbus.GetBus()
	_ = bus.Publish(context.Background(), "leaderboard:update", &model.ScoreUpdate{
		GameID:  gameID,
		Entries: entries,
	})

	return nil
}
