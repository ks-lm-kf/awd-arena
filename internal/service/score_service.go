package service

import (
	"encoding/json"

	"github.com/awd-platform/awd-arena/internal/model"
	"gorm.io/gorm"
)

// ScoreService 分数服务
type ScoreService struct {
	db *gorm.DB
}

// NewScoreService 创建ScoreService
func NewScoreService(db *gorm.DB) *ScoreService {
	return &ScoreService{db: db}
}

// UserContainerInfo 用户容器信息（返回给前端）
type UserContainerInfo struct {
	ContainerID   string `json:"container_id"`
	IPAddress     string `json:"ip_address"`
	SSHUser       string `json:"ssh_user"`
	SSHPassword   string `json:"ssh_password"`
	Ports         []int  `json:"ports"`
	ChallengeName string `json:"challenge_name"`
}

// AdjustScore 调整分数
func (s *ScoreService) AdjustScore(gameID, teamID int64, adjustValue int, reason string, operatorID int64) (*model.ScoreAdjustment, error) {
	// 获取当前比赛轮次
	var game model.Game
	if err := s.db.First(&game, gameID).Error; err != nil {
		return nil, err
	}

	// 创建调整记录
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

	// TODO: 更新队伍总分（需要在比赛计分逻辑中实现）
	// 这里只记录调整，实际分数计算在计分引擎中处理

	return adjustment, nil
}

// GetAdjustments 获取调整历史
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

// GetUserContainers 获取用户的容器信息
func (s *ScoreService) GetUserContainers(userID, gameID int64) ([]UserContainerInfo, error) {
	// 先找到用户对应的队伍
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}

	if user.TeamID == nil {
		return nil, gorm.ErrRecordNotFound
	}

	// 查询该队伍在该比赛中的所有容器
	var containers []model.TeamContainer
	err := s.db.Where("team_id = ? AND game_id = ?", *user.TeamID, gameID).
		Find(&containers).Error
	if err != nil {
		return nil, err
	}

	// 查询所有相关的 Challenge 信息
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

	// 转换为返回格式
	result := make([]UserContainerInfo, len(containers))
	for i, c := range containers {
		// 解析端口映射
		var ports []int
		if c.PortMapping != "" {
			// PortMapping 可能是 JSON 格式，例如 {"22": 31000, "80": 31001}
			var portMap map[string]int
			if err := json.Unmarshal([]byte(c.PortMapping), &portMap); err == nil {
				for port := range portMap {
					// 将字符串端口转换为整数
					var p int
					if _, err := json.Number(port).Int64(); err == nil {
						json.Unmarshal([]byte(port), &p)
						ports = append(ports, p)
					}
				}
			}
		}

		// 获取 Challenge 名称
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

