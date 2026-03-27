package notify

import (
	"fmt"
	"time"
)

// Example 使用示例
func Example() {
	// 1. 创建告警管理器
	alertManager := NewAlertManager()
	
	// 2. 配置飞书通知器
	feishuWebhook := "https://open.feishu.cn/open-apis/bot/v2/hook/your-webhook-token"
	feishuNotifier := NewFeishuNotifier(feishuWebhook)
	alertManager.AddNotifier(feishuNotifier)
	
	// 3. 配置邮件通知器（可选）
	emailConfig := EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		Username: "your-email@example.com",
		Password: "your-password",
		From:     "your-email@example.com",
		To:       []string{"admin@example.com", "ops@example.com"},
		UseTLS:   true,
	}
	emailNotifier := NewEmailNotifier(emailConfig)
	alertManager.AddNotifier(emailNotifier)
	
	// 4. 配置 WebSocket 通知器（可选）
	wsConfig := WebSocketConfig{
		URL:               "ws://localhost:8080/alerts",
		ReconnectInterval: 5 * time.Second,
		PingInterval:      30 * time.Second,
		WriteTimeout:      10 * time.Second,
	}
	wsNotifier := NewWebSocketNotifier(wsConfig)
	alertManager.AddNotifier(wsNotifier)
	
	// 5. 配置静默时间（相同告警在静默期内不重复发送）
	alertManager.SetSilence(AlertTypeServiceDown, 10*time.Minute)
	alertManager.SetSilence(AlertTypeDDoSAttack, 5*time.Minute)
	alertManager.SetSilence(AlertTypeHighCPU, 15*time.Minute)
	
	// 6. 发送服务宕机告警
	alert := CreateServiceDownAlert(
		"awd-arena-server",
		"服务进程异常退出，退出码: 1\n最后日志: Connection timeout",
	)
	if err := alertManager.SendAlert(alert); err != nil {
		fmt.Printf("发送告警失败: %v\n", err)
	}
	
	// 7. 发送 DDoS 攻击告警
	ddosAlert := CreateDDoSAlert(
		"192.168.1.100",
		50000,
		"检测到异常流量，每秒包数超过阈值",
	)
	if err := alertManager.SendAlert(ddosAlert); err != nil {
		fmt.Printf("发送告警失败: %v\n", err)
	}
	
	// 8. 发送高 CPU 告警
	cpuAlert := CreateHighCPUAlert(92.5)
	if err := alertManager.SendAlert(cpuAlert); err != nil {
		fmt.Printf("发送告警失败: %v\n", err)
	}
	
	// 9. 发送高内存告警
	memAlert := CreateHighMemoryAlert(88.3)
	if err := alertManager.SendAlert(memAlert); err != nil {
		fmt.Printf("发送告警失败: %v\n", err)
	}
	
	// 10. 测试飞书连接（可选）
	if err := feishuNotifier.SendSimpleText("告警系统已启动"); err != nil {
		fmt.Printf("飞书连接测试失败: %v\n", err)
	}
	
	fmt.Println("告警通知系统示例完成")
}
