package model

import "time"

// Health status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

type ServiceHealth struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	ServiceID    uint      `json:"service_id" gorm:"index"`
	Status       string    `json:"status"`
	CheckedAt    time.Time `json:"checked_at"`
	ResponseTime int64     `json:"response_time"`
	ErrorMsg     string    `json:"error_msg"`
	Notified     bool      `json:"notified"`
	CreatedAt    time.Time `json:"created_at"`
}

type HealthCheckConfig struct {
	Interval       int    `json:"interval"`
	Timeout        int    `json:"timeout"`
	MaxRetries     int    `json:"max_retries"`
	CheckInterval  int    `json:"check_interval"`
	FailureCount   int    `json:"failure_count"`
	RecoveryNotify bool   `json:"recovery_notify"`
	HTTPPath       string `json:"http_path"`
	ExpectedCode   int    `json:"expected_code"`
}

// DefaultHealthCheckConfig returns a default health check config.
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		Interval:       30,
		Timeout:        10,
		MaxRetries:     3,
		CheckInterval:  30,
		FailureCount:   3,
		RecoveryNotify: true,
		HTTPPath:       "/health",
		ExpectedCode:   200,
	}
}

// ServiceHealthSummary holds aggregated health stats.
type ServiceHealthSummary struct {
	ServiceID       uint      `json:"service_id"`
	ServiceName     string    `json:"service_name"`
	CurrentStatus   string    `json:"current_status"`
	LastCheckedAt   time.Time `json:"last_checked_at"`
	UptimePercent   float64   `json:"uptime_percent"`
	TotalChecks     int64     `json:"total_checks"`
	HealthyChecks   int64     `json:"healthy_checks"`
	AvgResponseTime int64     `json:"avg_response_time"`
}
