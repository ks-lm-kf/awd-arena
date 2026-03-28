package model

type Ranking struct {
    Rank        uint    `json:"rank"`
    TeamID      uint    `json:"team_id"`
    TeamName    string  `json:"team_name"`
    Score       float64 `json:"score"`
    Attacks     int     `json:"attacks"`
    Defenses    int     `json:"defenses"`
    FirstBlood  int     `json:"first_blood"`
}

type RoundRanking struct {
    Round    int       `json:"round"`
    Rankings []Ranking `json:"rankings"`
}

type CompetitionStats struct {
    TotalTeams       int     `json:"total_teams"`
    TotalRounds      int     `json:"total_rounds"`
    TotalAttacks     int     `json:"total_attacks"`
    AvgAttacksPerRound float64 `json:"avg_attacks_per_round"`
    MostAttackedTeam string  `json:"most_attacked_team"`
    TopAttacker      string  `json:"top_attacker"`
}
