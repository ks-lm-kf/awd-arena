# 告警通知系统

## 概述

告警通知系统支持多渠道告警，包括：
- **飞书**：通过 Webhook 发送富文本消息卡片
- **邮件**：发送 HTML 格式的告警邮件
- **WebSocket**：实时推送告警到前端

## 核心功能

### 1. 多渠道支持
- 飞书 Webhook（已实现）
- SMTP 邮件（已实现）
- WebSocket 实时推送（已实现）

### 2. 告警去重
- 基于告警内容（类型+级别+标题+消息）生成唯一 ID
- 相同告警在静默期内不会重复发送

### 3. 静默时间
- 可为每类告警配置不同的静默时长
- 默认静默时间：5 分钟
- 示例：`alertManager.SetSilence(AlertTypeServiceDown, 10*time.Minute)`

### 4. 告警级别
- `INFO`：普通信息（蓝色）
- `WARNING`：警告（橙色）
- `CRITICAL`：严重告警（红色）

### 5. 告警类型
- `service_down`：服务宕机
- `ddos_attack`：DDoS 攻击
- `high_cpu`：高 CPU 使用率
- `high_memory`：高内存使用率
- `network_error`：网络错误

## 快速开始

### 基本使用

```go
package main

import (
    "time"
    "your-project/internal/notify"
)

func main() {
    // 1. 创建告警管理器
    alertManager := notify.NewAlertManager()
    
    // 2. 添加飞书通知器
    feishu := notify.NewFeishuNotifier("https://open.feishu.cn/open-apis/bot/v2/hook/your-token")
    alertManager.AddNotifier(feishu)
    
    // 3. 配置静默时间
    alertManager.SetSilence(notify.AlertTypeServiceDown, 10*time.Minute)
    
    // 4. 发送告警
    alert := notify.CreateServiceDownAlert("my-service", "进程异常退出")
    alertManager.SendAlert(alert)
}
```

### 飞书配置

1. 在飞书群组中添加自定义机器人
2. 获取 Webhook URL
3. 创建通知器：

```go
feishu := notify.NewFeishuNotifier("your-webhook-url")
alertManager.AddNotifier(feishu)
```

### 邮件配置

```go
emailConfig := notify.EmailConfig{
    SMTPHost: "smtp.gmail.com",
    SMTPPort: 587,
    Username: "your-email@gmail.com",
    Password: "your-app-password",
    From:     "your-email@gmail.com",
    To:       []string{"admin@company.com"},
    UseTLS:   true,
}
email := notify.NewEmailNotifier(emailConfig)
alertManager.AddNotifier(email)
```

### WebSocket 配置

```go
wsConfig := notify.WebSocketConfig{
    URL:               "ws://localhost:8080/alerts",
    ReconnectInterval: 5 * time.Second,
    PingInterval:      30 * time.Second,
}
ws := notify.NewWebSocketNotifier(wsConfig)
alertManager.AddNotifier(ws)
```

## API 文档

### AlertManager

#### `NewAlertManager() *AlertManager`
创建告警管理器实例

#### `AddNotifier(notifier Notifier)`
添加通知器（飞书/邮件/WebSocket）

#### `SetSilence(alertType AlertType, duration time.Duration)`
设置某类告警的静默时间

#### `SendAlert(alert Alert) error`
发送告警（自动去重和静默）

### 辅助函数

#### `CreateServiceDownAlert(serviceName, message string) Alert`
创建服务宕机告警

#### `CreateDDoSAlert(source string, pps int, message string) Alert`
创建 DDoS 攻击告警

#### `CreateHighCPUAlert(usage float64) Alert`
创建高 CPU 告警

#### `CreateHighMemoryAlert(usage float64) Alert`
创建高内存告警

## 配置示例

### config.yaml

```yaml
alert:
  enabled: true
  dedupe: true
  silence:
    service_down: 10m
    ddos_attack: 5m
    high_cpu: 15m
    high_memory: 15m

feishu:
  webhook: "https://open.feishu.cn/open-apis/bot/v2/hook/your-token"
  timeout: 10s

email:
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "your-email@gmail.com"
  password: "your-app-password"
  from: "your-email@gmail.com"
  to:
    - "admin@company.com"
    - "ops@company.com"
  use_tls: true

websocket:
  url: "ws://localhost:8080/alerts"
  reconnect_interval: 5s
  ping_interval: 30s
  write_timeout: 10s
```

## 集成到项目

### 在主程序中初始化

```go
// cmd/server/main.go

func setupAlertManager(cfg *config.Config) *notify.AlertManager {
    am := notify.NewAlertManager()
    
    // 飞书
    if cfg.Alert.Feishu.Webhook != "" {
        feishu := notify.NewFeishuNotifier(cfg.Alert.Feishu.Webhook)
        am.AddNotifier(feishu)
    }
    
    // 邮件
    if cfg.Alert.Email.SMTPHost != "" {
        emailConfig := notify.EmailConfig{
            SMTPHost: cfg.Alert.Email.SMTPHost,
            SMTPPort: cfg.Alert.Email.SMTPPort,
            Username: cfg.Alert.Email.Username,
            Password: cfg.Alert.Email.Password,
            From:     cfg.Alert.Email.From,
            To:       cfg.Alert.Email.To,
            UseTLS:   cfg.Alert.Email.UseTLS,
        }
        email := notify.NewEmailNotifier(emailConfig)
        am.AddNotifier(email)
    }
    
    // 配置静默
    for alertType, duration := range cfg.Alert.Silence {
        am.SetSilence(alertType, duration)
    }
    
    return am
}
```

### 在监控模块中使用

```go
// internal/monitor/service.go

func (m *Monitor) CheckServiceHealth() {
    for {
        if !m.service.IsRunning() {
            alert := notify.CreateServiceDownAlert(
                m.service.Name,
                "服务进程已停止",
            )
            m.alertManager.SendAlert(alert)
        }
        time.Sleep(30 * time.Second)
    }
}
```

## 测试

### 测试飞书连接

```go
feishu := notify.NewFeishuNotifier("your-webhook-url")
err := feishu.SendSimpleText("测试消息")
if err != nil {
    log.Fatal("飞书连接失败:", err)
}
```

### 测试邮件发送

```go
email := notify.NewEmailNotifier(emailConfig)
err := email.SendSimpleText("测试主题", "这是一封测试邮件")
if err != nil {
    log.Fatal("邮件发送失败:", err)
}
```

## 注意事项

1. **飞书 Webhook 限制**：每个机器人每分钟最多发送 20 条消息
2. **邮件发送频率**：建议配置合理的静默时间，避免邮件轰炸
3. **WebSocket 连接**：自动重连机制确保连接可靠性
4. **去重机制**：基于内容哈希，相同内容的告警会被去重
5. **静默时间**：在静默期内，相同告警不会重复发送

## 故障排查

### 飞书发送失败
- 检查 Webhook URL 是否正确
- 检查网络连接
- 查看飞书机器人是否被禁用

### 邮件发送失败
- 检查 SMTP 配置
- 检查用户名和密码
- 检查 TLS 配置
- 某些邮箱需要使用应用专用密码

### WebSocket 连接失败
- 检查服务器 URL 是否正确
- 检查服务器是否运行
- 查看防火墙设置

## 许可证

MIT License
