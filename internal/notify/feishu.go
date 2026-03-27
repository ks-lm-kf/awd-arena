package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FeishuConfig 飞书配置
type FeishuConfig struct {
	WebhookURL string // 飞书机器人 Webhook URL
	Timeout    time.Duration
}

// FeishuNotifier 飞书通知器
type FeishuNotifier struct {
	config FeishuConfig
	client *http.Client
}

// NewFeishuNotifier 创建飞书通知器
func NewFeishuNotifier(webhookURL string) *FeishuNotifier {
	return &FeishuNotifier{
		config: FeishuConfig{
			WebhookURL: webhookURL,
			Timeout:    10 * time.Second,
		},
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// NewFeishuNotifierWithConfig 创建带配置的飞书通知器
func NewFeishuNotifierWithConfig(config FeishuConfig) *FeishuNotifier {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	return &FeishuNotifier{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// Name 返回通知器名称
func (fn *FeishuNotifier) Name() string {
	return "feishu"
}

// FeishuMessage 飞书消息结构
type FeishuMessage struct {
	MsgType string                 `json:"msg_type"`
	Content map[string]interface{} `json:"content"`
}

// FeishuCard 飞书消息卡片
type FeishuCard struct {
	MsgType string `json:"msg_type"`
	Card    Card   `json:"card"`
}

// Card 消息卡片
type Card struct {
	Config   CardConfig   `json:"config"`
	Header   CardHeader   `json:"header"`
	Elements []CardElement `json:"elements"`
}

// CardConfig 卡片配置
type CardConfig struct {
	WideScreenMode bool `json:"wide_screen_mode"`
	EnableForward  bool `json:"enable_forward"`
}

// CardHeader 卡片头部
type CardHeader struct {
	Title    CardTitle `json:"title"`
	Template string    `json:"template"`
}

// CardTitle 卡片标题
type CardTitle struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// CardElement 卡片元素
type CardElement struct {
	Tag  string `json:"tag"`
	Text CardText `json:"text,omitempty"`
}

// CardText 卡片文本
type CardText struct {
	Tag     string `json:"tag"`
	Content string `json:"content"`
}

// Send 发送告警到飞书
func (fn *FeishuNotifier) Send(alert Alert) error {
	// 构建消息卡片
	card := fn.buildCard(alert)
	
	// 发送请求
	return fn.sendCard(card)
}

// buildCard 构建消息卡片
func (fn *FeishuNotifier) buildCard(alert Alert) FeishuCard {
	// 根据告警级别选择颜色模板
	template := fn.getLevelTemplate(alert.Level)
	
	// 构建卡片内容
	elements := []CardElement{
		{
			Tag: "div",
			Text: CardText{
				Tag:     "lark_md",
				Content: fn.formatMessage(alert),
			},
		},
		{
			Tag: "div",
			Text: CardText{
				Tag:     "lark_md",
				Content: fmt.Sprintf("⏰ 时间: %s", alert.Timestamp.Format("2006-01-02 15:04:05")),
			},
		},
	}
	
	// 添加元数据
	if len(alert.Metadata) > 0 {
		metadataStr := "**附加信息:**\n"
		for k, v := range alert.Metadata {
			metadataStr += fmt.Sprintf("- %s: %v\n", k, v)
		}
		elements = append(elements, CardElement{
			Tag: "div",
			Text: CardText{
				Tag:     "lark_md",
				Content: metadataStr,
			},
		})
	}
	
	return FeishuCard{
		MsgType: "interactive",
		Card: Card{
			Config: CardConfig{
				WideScreenMode: true,
				EnableForward:  true,
			},
			Header: CardHeader{
				Title: CardTitle{
					Tag:     "plain_text",
					Content: alert.Title,
				},
				Template: template,
			},
			Elements: elements,
		},
	}
}

// getLevelTemplate 根据告警级别获取模板
func (fn *FeishuNotifier) getLevelTemplate(level AlertLevel) string {
	switch level {
	case AlertLevelInfo:
		return "blue"
	case AlertLevelWarning:
		return "orange"
	case AlertLevelCritical:
		return "red"
	default:
		return "grey"
	}
}

// formatMessage 格式化消息内容
func (fn *FeishuNotifier) formatMessage(alert Alert) string {
	levelEmoji := fn.getLevelEmoji(alert.Level)
	return fmt.Sprintf("%s **级别:** %s\n\n**类型:** %s\n\n**详情:**\n%s",
		levelEmoji,
		alert.Level.String(),
		alert.Type,
		alert.Message)
}

// getLevelEmoji 获取级别对应的emoji
func (fn *FeishuNotifier) getLevelEmoji(level AlertLevel) string {
	switch level {
	case AlertLevelInfo:
		return "ℹ️"
	case AlertLevelWarning:
		return "⚠️"
	case AlertLevelCritical:
		return "🚨"
	default:
		return "📢"
	}
}

// sendCard 发送卡片消息
func (fn *FeishuNotifier) sendCard(card FeishuCard) error {
	// 序列化消息
	body, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", fn.config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	resp, err := fn.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}
	
	// 检查状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}
	
	// 解析响应
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("parse response failed: %w", err)
	}
	
	if result.Code != 0 {
		return fmt.Errorf("feishu api error: code=%d, msg=%s", result.Code, result.Msg)
	}
	
	return nil
}

// SendSimpleText 发送简单文本消息（用于测试）
func (fn *FeishuNotifier) SendSimpleText(text string) error {
	msg := FeishuMessage{
		MsgType: "text",
		Content: map[string]interface{}{
			"text": text,
		},
	}
	
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal message failed: %w", err)
	}
	
	req, err := http.NewRequest("POST", fn.config.WebhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request failed: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := fn.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request failed: %w", err)
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response failed: %w", err)
	}
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("feishu api error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}
	
	return nil
}
