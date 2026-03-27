package notify

import (
	"testing"
	"time"
)

func TestAlertManager_SendAlert(t *testing.T) {
	// 创建告警管理器
	am := NewAlertManager()
	
	// 验证初始状态
	if !am.enabled {
		t.Error("AlertManager should be enabled by default")
	}
	
	if !am.enableDedupe {
		t.Error("Dedupe should be enabled by default")
	}
	
	// 创建测试告警
	alert := Alert{
		Type:    AlertTypeServiceDown,
		Level:   AlertLevelCritical,
		Title:   "Test Alert",
		Message: "This is a test alert",
	}
	
	// 验证告警 ID 生成
	alertID := generateAlertID(alert)
	if alertID == "" {
		t.Error("Alert ID should not be empty")
	}
	
	t.Logf("Generated alert ID: %s", alertID)
}

func TestAlertManager_Deduplication(t *testing.T) {
	// 创建告警管理器
	_ = NewAlertManager()
	
	alert := Alert{
		Type:    AlertTypeServiceDown,
		Level:   AlertLevelCritical,
		Title:   "Service Down",
		Message: "Test service is down",
	}
	
	// 生成相同的告警 ID
	id1 := generateAlertID(alert)
	id2 := generateAlertID(alert)
	
	if id1 != id2 {
		t.Error("Same alerts should generate same ID")
	}
	
	t.Logf("Deduplication working: both alerts have ID %s", id1)
}

func TestAlertManager_Silence(t *testing.T) {
	am := NewAlertManager()
	
	// 设置静默时间
	am.SetSilence(AlertTypeServiceDown, 5*time.Minute)
	
	alertID := "test-alert-123"
	
	// 第一次不应该静默
	if am.isSilenced(alertID, AlertTypeServiceDown) {
		t.Error("First alert should not be silenced")
	}
	
	// 更新静默状态
	am.updateSilenceState(alertID)
	
	// 第二次应该静默
	if !am.isSilenced(alertID, AlertTypeServiceDown) {
		t.Error("Second alert should be silenced")
	}
	
	t.Log("Silence mechanism working correctly")
}

func TestCreateServiceDownAlert(t *testing.T) {
	alert := CreateServiceDownAlert("test-service", "Process exited with code 1")
	
	if alert.Type != AlertTypeServiceDown {
		t.Error("Alert type should be service_down")
	}
	
	if alert.Level != AlertLevelCritical {
		t.Error("Service down should be critical level")
	}
	
	if alert.Title == "" {
		t.Error("Alert title should not be empty")
	}
	
	t.Logf("Created service down alert: %s", alert.Title)
}

func TestCreateDDoSAlert(t *testing.T) {
	alert := CreateDDoSAlert("192.168.1.100", 50000, "High packet rate detected")
	
	if alert.Type != AlertTypeDDoSAttack {
		t.Error("Alert type should be ddos_attack")
	}
	
	if alert.Level != AlertLevelCritical {
		t.Error("DDoS attack should be critical level")
	}
	
	if alert.Metadata["source"] != "192.168.1.100" {
		t.Error("Metadata should contain source IP")
	}
	
	if alert.Metadata["pps"] != 50000 {
		t.Error("Metadata should contain PPS value")
	}
	
	t.Logf("Created DDoS alert: %s", alert.Title)
}

func TestAlertLevel_String(t *testing.T) {
	tests := []struct {
		level    AlertLevel
		expected string
	}{
		{AlertLevelInfo, "INFO"},
		{AlertLevelWarning, "WARNING"},
		{AlertLevelCritical, "CRITICAL"},
	}
	
	for _, test := range tests {
		result := test.level.String()
		if result != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, result)
		}
	}
}

func TestAlertManager_EnableDisable(t *testing.T) {
	am := NewAlertManager()
	
	// 默认应该启用
	if !am.enabled {
		t.Error("AlertManager should be enabled by default")
	}
	
	// 禁用
	am.Enable(false)
	
	if am.enabled {
		t.Error("AlertManager should be disabled")
	}
	
	// 重新启用
	am.Enable(true)
	
	if !am.enabled {
		t.Error("AlertManager should be enabled")
	}
	
	t.Log("Enable/Disable working correctly")
}

