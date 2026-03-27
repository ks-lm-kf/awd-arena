package model

import (
	"time"
)

// ServiceHealth 服务健康状态记录
type ServiceHealth struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ServiceID   uint      `gorm:"index;not null" json:"service_id"`     // 关联的服务ID
	Status      string    `gorm:"size:20;not null" json:"status"`        // healthy, unhealthy, unknown
	CheckedAt   time.Time `gorm:"index;not null" json:"checked_at"`      // 检查时间
	ResponseTime int64    `json:"response_time"`                         // 响应时间(毫秒)
	ErrorMsg    string    `gorm:"type:text" json:"error_msg"`            // 错误信息
	Notified    bool      `gorm:"default:false" json:"notified"`         // 是否已发送告警
	CreatedAt   time.Time `json:"created_at"`
}

// TableName 指定表名
func (ServiceHealth) TableName() string {
	return "service_health"
}

// HealthStatus 健康状态常量
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusUnknown   = "unknown"
)

// ServiceHealthSummary 服务健康摘要
type ServiceHealthSummary struct {
	ServiceID       uint    `json:"service_id"`
	ServiceName     string  `json:"service_name"`
	CurrentStatus   string  `json:"current_status"`
	LastCheckedAt   time.Time `json:"last_checked_at"`
	UptimePercent   float64 `json:"uptime_percent"` // 可用率
	TotalChecks     int64   `json:"total_checks"`
	HealthyChecks   int64   `json:"healthy_checks"`
	AvgResponseTime int64   `json:"avg_response_time_ms"`
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	CheckInterval   int           `yaml:"check_interval" json:"check_interval"`     // 检查间隔(秒), 默认10
	Timeout         int           `yaml:"timeout" json:"timeout"`                   // 超时时间(秒), 默认5
	MaxRetries      int           `yaml:"max_retries" json:"max_retries"`           // 最大重试次数, 默认3
	RetryInterval   int           `yaml:"retry_interval" json:"retry_interval"`     // 重试间隔(秒), 默认2
	FailureCount    int           `yaml:"failure_count" json:"failure_count"`       // 连续失败次数触发告警, 默认3
	RecoveryNotify  bool          `yaml:"recovery_notify" json:"recovery_notify"`   // 服务恢复时通知, 默认true
}

// DefaultHealthCheckConfig 默认健康检查配置
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		CheckInterval:  10,
		Timeout:        5,
		MaxRetries:     3,
		RetryInterval:  2,
		FailureCount:   3,
		RecoveryNotify: true,
	}
}
