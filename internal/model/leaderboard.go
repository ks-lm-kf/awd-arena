package model

// LeaderboardEntry represents a single entry in the leaderboard
type LeaderboardEntry struct {
	Rank         int     `json:"rank"`
	TeamID       int64   `json:"team_id"`
	TeamName     string  `json:"team_name"`
	TotalScore   float64 `json:"total_score"`
	AttackScore  float64 `json:"attack_score"`
	DefenseScore float64 `json:"defense_score"`
	FirstBloods  int     `json:"first_bloods"`
}

// ScoreUpdate represents a leaderboard update message
type ScoreUpdate struct {
	GameID  int64               `json:"game_id"`
	Entries []LeaderboardEntry `json:"entries"`
}
