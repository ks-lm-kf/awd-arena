package model

import "time"

type AttackLog struct {
    Timestamp    time.Time `json:"timestamp" gorm:"primaryKey"`
    GameID       uint64    `json:"game_id"`
    Round        uint32    `json:"round"`
    AttackerTeam string    `json:"attacker_team"`
    TargetTeam   string    `json:"target_team"`
    TargetIP     string    `json:"target_ip"`
    TargetPort   uint16    `json:"target_port"`
    Protocol     string    `json:"protocol"`
    Method       *string   `json:"method"`
    Path         *string   `json:"path"`
    PayloadHash  string    `json:"payload_hash"`
    AttackType   string    `json:"attack_type"`
    Severity     string    `json:"severity"`
    SourceIP     string    `json:"source_ip"`
    UserAgent    *string   `json:"user_agent"`
    RawLog       string    `json:"raw_log"`
}

type EventLog struct {
    ID        int64      `json:"id" gorm:"primaryKey"`
    GameID    *int64     `json:"game_id" gorm:"index"`
    EventType string     `json:"event_type"`
    Level     string     `json:"level"`
    TeamID    *int64     `json:"team_id"`
    Detail    string     `json:"detail"`
    CreatedAt time.Time  `json:"created_at"`
}
