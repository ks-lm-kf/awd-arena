package handler

import (
	"context"
	"sort"

	"github.com/awd-platform/awd-arena/internal/database"
	"github.com/awd-platform/awd-arena/internal/eventbus"
	"github.com/awd-platform/awd-arena/internal/model"
	"github.com/gofiber/fiber/v3"
	"gorm.io/gorm"
)

var LeaderboardHandler *leaderboardHandler

func init() {
	LeaderboardHandler = &leaderboardHandler{}
}

type leaderboardHandler struct{}

type RankHistoryEntry struct {
	Round      int     `json:"round"`
	TeamID     int64   `json:"team_id"`
	TeamName   string  `json:"team_name"`
	Rank       int     `json:"rank"`
	TotalScore float64 `json:"total_score"`
}

type LeaderboardResponse struct {
	GameID  int64                    `json:"game_id"`
	Entries []model.LeaderboardEntry `json:"entries"`
}

// Get returns the current leaderboard for a game
func (h *leaderboardHandler) Get(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}
	entries, err := h.buildLeaderboard(db, gameID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": entries})
}

// GetRound returns rankings for a specific round
func (h *leaderboardHandler) GetRound(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	round, _ := parseID(c.Params("round"))
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}
	teamMap := h.getTeamMap(db)
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ? AND round = ?", gameID, int(round)).Order("total_score desc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	var entries []model.LeaderboardEntry
	for i, rs := range roundScores {
		entries = append(entries, model.LeaderboardEntry{
			TeamID: rs.TeamID, TeamName: teamMap[rs.TeamID],
			TotalScore: rs.TotalScore, AttackScore: rs.AttackScore, DefenseScore: rs.DefenseScore, Rank: i + 1,
		})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": entries})
}

// GetHistory returns historical rankings
func (h *leaderboardHandler) GetHistory(c fiber.Ctx) error {
	gameID, err := parseID(c.Params("id"))
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"code": 400, "message": err.Error()})
	}
	db := database.GetDB()
	if db == nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": "database not available"})
	}
	teamMap := h.getTeamMap(db)
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Order("round, total_score desc").Find(&roundScores).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"code": 500, "message": err.Error()})
	}
	var history []RankHistoryEntry
	for _, rs := range roundScores {
		history = append(history, RankHistoryEntry{
			Round: rs.Round, TeamID: rs.TeamID, TeamName: teamMap[rs.TeamID],
			TotalScore: rs.TotalScore, Rank: rs.Rank,
		})
	}
	return c.JSON(fiber.Map{"code": 0, "message": "ok", "data": history})
}

// BroadcastUpdate rebuilds and broadcasts leaderboard
func (h *leaderboardHandler) BroadcastUpdate(gameID int64, db *gorm.DB) error {
	entries, err := h.buildLeaderboard(db, gameID)
	if err != nil {
		return err
	}
	bus := eventbus.GetBus()
	if bus != nil {
		bus.Publish(context.Background(), "ranking:update", entries)
	}
	return nil
}

func (h *leaderboardHandler) buildLeaderboard(db *gorm.DB, gameID int64) ([]model.LeaderboardEntry, error) {
	teamMap := h.getTeamMap(db)
	var roundScores []model.RoundScore
	if err := db.Where("game_id = ?", gameID).Find(&roundScores).Error; err != nil {
		return nil, err
	}
	type teamAgg struct {
		TeamID       int64
		TotalScore   float64
		AttackScore  float64
		DefenseScore float64
	}
	aggMap := make(map[int64]*teamAgg)
	for _, rs := range roundScores {
		if _, ok := aggMap[rs.TeamID]; !ok {
			aggMap[rs.TeamID] = &teamAgg{TeamID: rs.TeamID}
		}
		aggMap[rs.TeamID].TotalScore += rs.TotalScore
		aggMap[rs.TeamID].AttackScore += rs.AttackScore
		aggMap[rs.TeamID].DefenseScore += rs.DefenseScore
	}
	var entries []model.LeaderboardEntry
	for _, agg := range aggMap {
		entries = append(entries, model.LeaderboardEntry{
			TeamID: agg.TeamID, TeamName: teamMap[agg.TeamID],
			TotalScore: agg.TotalScore, AttackScore: agg.AttackScore, DefenseScore: agg.DefenseScore,
		})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].TotalScore > entries[j].TotalScore })
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries, nil
}

func (h *leaderboardHandler) getTeamMap(db *gorm.DB) map[int64]string {
	var teams []model.Team
	db.Find(&teams)
	m := make(map[int64]string, len(teams))
	for _, t := range teams {
		m[t.ID] = t.Name
	}
	return m
}
