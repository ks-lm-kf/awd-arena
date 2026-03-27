package notify

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// AlertLevel 告警级别
type AlertLevel int

const (
	AlertLevelInfo AlertLevel = iota
	AlertLevelWarning
	AlertLevelCritical
)

func (l AlertLevel) String() string {
	switch l {
	case AlertLevelInfo:
		return "INFO"
	case AlertLevelWarning:
		return "WARNING"
	case AlertLevelCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// AlertType 告警类型
type AlertType string

const (
	AlertTypeServiceDown  AlertType = "service_down"
	AlertTypeDDoSAttack   AlertType = "ddos_attack"
	AlertTypeHighCPU      AlertType = "high_cpu"
	AlertTypeHighMemory   AlertType = "high_memory"
	AlertTypeNetworkError AlertType = "network_error"
)

// Alert 告警信息
type Alert struct {
	ID        string                 `json:"id"`         // 告警唯一标识（用于去重）
	Type      AlertType              `json:"type"`       // 告警类型
	Level     AlertLevel             `json:"level"`      // 告警级别
	Title     string                 `json:"title"`      // 告警标题
	Message   string                 `json:"message"`    // 告警详细消息
	Timestamp time.Time              `json:"timestamp"`  // 告警时间
	Metadata  map[string]interface{} `json:"metadata"`   // 附加元数据
}

// Notifier 告警通知器接口
type Notifier interface {
	Name() string
	Send(alert Alert) error
}

// SilenceConfig 静默配置
type SilenceConfig struct {
	Duration time.Duration // 静默时长
}

// AlertManager 告警管理器
type AlertManager struct {
	mu               sync.RWMutex
	notifiers        []Notifier
	silenceConfigs   map[AlertType]SilenceConfig
	silenceState     map[string]time.Time // 记录每个告警的最后发送时间
	enabled          bool
	enableDedupe     bool // 是否启用去重
	enableSilence    bool // 是否启用静默
}

// NewAlertManager 创建告警管理器
func NewAlertManager() *AlertManager {
	return &AlertManager{
		notifiers:      make([]Notifier, 0),
		silenceConfigs: make(map[AlertType]SilenceConfig),
		silenceState:   make(map[string]time.Time),
		enabled:        true,
		enableDedupe:   true,
		enableSilence:  true,
	}
}

// AddNotifier 添加通知器
func (am *AlertManager) AddNotifier(notifier Notifier) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.notifiers = append(am.notifiers, notifier)
}

// SetSilence 设置某类告警的静默时间
func (am *AlertManager) SetSilence(alertType AlertType, duration time.Duration) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.silenceConfigs[alertType] = SilenceConfig{Duration: duration}
}

// Enable 启用/禁用告警
func (am *AlertManager) Enable(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.enabled = enabled
}

// SetDedupe 设置是否启用去重
func (am *AlertManager) SetDedupe(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.enableDedupe = enabled
}

// SetSilenceEnabled 设置是否启用静默
func (am *AlertManager) SetSilenceEnabled(enabled bool) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.enableSilence = enabled
}

// generateAlertID 生成告警ID（用于去重）
func generateAlertID(alert Alert) string {
	// 基于告警类型、级别和消息内容生成唯一ID
	hash := sha256.New()
	hash.Write([]byte(string(alert.Type)))
	hash.Write([]byte(alert.Level.String()))
	hash.Write([]byte(alert.Title))
	hash.Write([]byte(alert.Message))
	return hex.EncodeToString(hash.Sum(nil))[:16]
}

// isSilenced 检查告警是否在静默期
func (am *AlertManager) isSilenced(alertID string, alertType AlertType) bool {
	am.mu.RLock()
	defer am.mu.RUnlock()

	if !am.enableSilence {
		return false
	}

	lastSent, exists := am.silenceState[alertID]
	if !exists {
		return false
	}

	config, hasConfig := am.silenceConfigs[alertType]
	if !hasConfig {
		// 默认静默5分钟
		config = SilenceConfig{Duration: 5 * time.Minute}
	}

	return time.Since(lastSent) < config.Duration
}

// updateSilenceState 更新静默状态
func (am *AlertManager) updateSilenceState(alertID string) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.silenceState[alertID] = time.Now()
}

// SendAlert 发送告警
func (am *AlertManager) SendAlert(alert Alert) error {
	am.mu.RLock()
	enabled := am.enabled
	enableDedupe := am.enableDedupe
	am.mu.RUnlock()

	if !enabled {
		return fmt.Errorf("alert manager is disabled")
	}

	// 生成告警ID
	alertID := generateAlertID(alert)
	alert.ID = alertID

	// 去重检查
	if enableDedupe && am.isSilenced(alertID, alert.Type) {
		return fmt.Errorf("alert is silenced (ID: %s)", alertID)
	}

	// 设置时间戳
	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	// 通过所有通知器发送
	am.mu.RLock()
	notifiers := am.notifiers
	am.mu.RUnlock()

	var lastErr error
	for _, notifier := range notifiers {
		if err := notifier.Send(alert); err != nil {
			lastErr = fmt.Errorf("notifier %s failed: %w", notifier.Name(), err)
		}
	}

	// 更新静默状态
	if lastErr == nil {
		am.updateSilenceState(alertID)
	}

	return lastErr
}

// CreateServiceDownAlert 创建服务宕机告警
func CreateServiceDownAlert(serviceName, message string) Alert {
	return Alert{
		Type:    AlertTypeServiceDown,
		Level:   AlertLevelCritical,
		Title:   fmt.Sprintf("服务宕机: %s", serviceName),
		Message: message,
	}
}

// CreateDDoSAlert 创建DDoS攻击告警
func CreateDDoSAlert(source string, pps int, message string) Alert {
	return Alert{
		Type:  AlertTypeDDoSAttack,
		Level: AlertLevelCritical,
		Title: "检测到异常攻击",
		Message: fmt.Sprintf("来源: %s\n包数: %d pps\n详情: %s", 
			source, pps, message),
		Metadata: map[string]interface{}{
			"source": source,
			"pps":    pps,
		},
	}
}

// CreateHighCPUAlert 创建高CPU告警
func CreateHighCPUAlert(usage float64) Alert {
	return Alert{
		Type:  AlertTypeHighCPU,
		Level: AlertLevelWarning,
		Title: "CPU使用率过高",
		Message: fmt.Sprintf("当前CPU使用率: %.2f%%", usage),
		Metadata: map[string]interface{}{
			"usage": usage,
		},
	}
}

// CreateHighMemoryAlert 创建高内存告警
func CreateHighMemoryAlert(usage float64) Alert {
	return Alert{
		Type:  AlertTypeHighMemory,
		Level: AlertLevelWarning,
		Title: "内存使用率过高",
		Message: fmt.Sprintf("当前内存使用率: %.2f%%", usage),
		Metadata: map[string]interface{}{
			"usage": usage,
		},
	}
}

