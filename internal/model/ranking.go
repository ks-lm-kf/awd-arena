package model

import "time"

// Ranking 排名信息
type Ranking struct {
    Rank       int       `json:rank`
    TeamID     uint64    `json:team_id`
    TeamName   string    `json:team_name`
    Score      float64   `json:score`
    Attacks    int       `json:attacks`
    Defenses   int       `json:defenses`
    FirstBlood int       `json:first_blood`
    UpdatedAt  time.Time `json:updated_at`
}


// RoundRanking 轮次排名
type RoundRanking struct {
    Round      uint32    `json:round`
    TeamID     uint64    `json:team_id`
    TeamName   string    `json:team_name`
    Score      float64   `json:score`
    UpdatedAt  time.Time `json:updated_at`
}

// CompetitionStats 比赛统计
type CompetitionStats struct {
    TotalTeams      int     `json:total_teams`
    TotalAttacks    int     `json:total_attacks`
    AvgScore        float64 `json:avg_score`
    TopTeam         string  `json:top_team`
    TopScore        float64 `json:top_score`
}
