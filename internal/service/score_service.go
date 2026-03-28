package service

import (
	"encoding/json"

	"github.com/awd-platform/awd-arena/internal/model"
	"gorm.io/gorm"
)

// ScoreService provides score-related operations.
type ScoreService struct {
	db *gorm.DB
}

// NewScoreService creates a ScoreService.
func NewScoreService(db *gorm.DB) *ScoreService {
	return &ScoreService{db: db}
}

// UserContainerInfo represents container info for the frontend.
type UserContainerInfo struct {
	ContainerID   string `json:"container_id"`
	IPAddress     string `json:"ip_address"`
	SSHUser       string `json:"ssh_user"`
	SSHPassword   string `json:"ssh_password"`
	Ports         []int  `json:"ports"`
	ChallengeName string `json:"challenge_name"`
}

// AdjustScore adjusts a team's score and updates the team's cumulative total.
func (s *ScoreService) AdjustScore(gameID, teamID int64, adjustValue int, reason string, operatorID int64) (*model.ScoreAdjustment, error) {
	var game model.Game
	if err := s.db.First(&game, gameID).Error; err != nil {
		return nil, err
	}

	adjustment := &model.ScoreAdjustment{
		GameID:      gameID,
		TeamID:      teamID,
		AdjustValue: adjustValue,
		Reason:      reason,
		OperatorID:  operatorID,
		Round:       game.CurrentRound,
	}

	if err := s.db.Create(adjustment).Error; err != nil {
		return nil, err
	}

	// Update team's cumulative score
	var currentScore float64
	s.db.Model(&model.Team{}).Where("id = ?", teamID).
		Select("COALESCE(score, 0)").Row().Scan(&currentScore)

	newScore := currentScore + float64(adjustValue)
	if err := s.db.Model(&model.Team{}).Where("id = ?", teamID).Update("score", newScore).Error; err != nil {
		return nil, err
	}

	return adjustment, nil
}

// GetAdjustments returns score adjustment history.
func (s *ScoreService) GetAdjustments(gameID, teamID int64) ([]model.ScoreAdjustment, error) {
	var adjustments []model.ScoreAdjustment
	query := s.db.Model(&model.ScoreAdjustment{})
	if gameID > 0 {
		query = query.Where("game_id = ?", gameID)
	}
	if teamID > 0 {
		query = query.Where("team_id = ?", teamID)
	}
	err := query.Order("created_at desc").Find(&adjustments).Error
	return adjustments, err
}

// GetUserContainers returns container info for a user in a game.
func (s *ScoreService) GetUserContainers(userID, gameID int64) ([]UserContainerInfo, error) {
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if user.TeamID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	var containers []model.TeamContainer
	err := s.db.Where("team_id = ? AND game_id = ?", *user.TeamID, gameID).
		Find(&containers).Error
	if err != nil {
		return nil, err
	}

	challengeIDs := make([]int64, len(containers))
	for i, c := range containers {
		challengeIDs[i] = c.ChallengeID
	}

	var challenges []model.Challenge
	challengeMap := make(map[int64]model.Challenge)
	if err := s.db.Where("id IN ?", challengeIDs).Find(&challenges).Error; err == nil {
		for _, chal := range challenges {
			challengeMap[chal.ID] = chal
		}
	}

	result := make([]UserContainerInfo, len(containers))
	for i, c := range containers {
		var ports []int
		if c.PortMapping != "" {
			var portMap map[string]int
			if err := json.Unmarshal([]byte(c.PortMapping), &portMap); err == nil {
				for port := range portMap {
					var p int
					if _, err := json.Number(port).Int64(); err == nil {
						json.Unmarshal([]byte(port), &p)
						ports = append(ports, p)
					}
				}
			}
		}

		challengeName := ""
		if chal, ok := challengeMap[c.ChallengeID]; ok {
			challengeName = chal.Name
		}

		result[i] = UserContainerInfo{
			ContainerID:   c.ContainerID,
			IPAddress:     c.IPAddress,
			SSHUser:       c.SSHUser,
			SSHPassword:   c.SSHPassword,
			Ports:         ports,
			ChallengeName: challengeName,
		}
	}

	return result, nil
}

// GetRoundScores returns all scores for a specific round in a game.
func (s *ScoreService) GetRoundScores(gameID, round int) ([]model.RoundScore, error) {
	var scores []model.RoundScore
	err := s.db.Where("game_id = ? AND round = ?", gameID, round).
		Order("rank asc, total_score desc").Find(&scores).Error
	return scores, err
}

// GetTeamGameScores returns all round scores for a team in a game.
func (s *ScoreService) GetTeamGameScores(gameID, teamID int64) ([]model.RoundScore, error) {
	var scores []model.RoundScore
	err := s.db.Where("game_id = ? AND team_id = ?", gameID, teamID).
		Order("round asc").Find(&scores).Error
	return scores, err
}
