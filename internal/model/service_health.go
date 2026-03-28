package model

import "time"

type ServiceHealth struct {
    ID           uint      `json:"id" gorm:"primaryKey"`
    ServiceID    uint      `json:"service_id" gorm:"index"`
    Status       string    `json:"status"`
    CheckedAt    time.Time `json:"checked_at"`
    ResponseTime int64     `json:"response_time"`
    ErrorMsg     string    `json:"error_msg"`
    Notified     bool      `json:"notified"`
}

type HealthCheckConfig struct {
    Interval    int    `json:"interval"`
    Timeout     int    `json:"timeout"`
    MaxRetries  int    `json:"max_retries"`
    HTTPPath    string `json:"http_path"`
    ExpectedCode int   `json:"expected_code"`
}
