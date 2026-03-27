package notify

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string // SMTP 服务器地址
	SMTPPort     int    // SMTP 端口
	Username     string // 用户名
	Password     string // 密码
	From         string // 发件人地址
	To           []string // 收件人列表
	UseTLS       bool   // 是否使用 TLS
}

// EmailNotifier 邮件通知器
type EmailNotifier struct {
	config EmailConfig
	auth   smtp.Auth
}

// NewEmailNotifier 创建邮件通知器
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	auth := smtp.PlainAuth("", config.Username, config.Password, config.SMTPHost)
	return &EmailNotifier{
		config: config,
		auth:   auth,
	}
}

// Name 返回通知器名称
func (en *EmailNotifier) Name() string {
	return "email"
}

// Send 发送告警邮件
func (en *EmailNotifier) Send(alert Alert) error {
	// 构建邮件内容
	subject := fmt.Sprintf("[%s] %s", alert.Level.String(), alert.Title)
	body := en.buildEmailBody(alert)
	
	// 构建邮件消息
	message := en.buildMessage(subject, body)
	
	// 发送邮件
	addr := fmt.Sprintf("%s:%d", en.config.SMTPHost, en.config.SMTPPort)
	
	if err := smtp.SendMail(addr, en.auth, en.config.From, en.config.To, []byte(message)); err != nil {
		return fmt.Errorf("send email failed: %w", err)
	}
	
	return nil
}

// buildMessage 构建邮件消息
func (en *EmailNotifier) buildMessage(subject, body string) string {
	msg := fmt.Sprintf("From: %s\r\n", en.config.From)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(en.config.To, ","))
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-version: 1.0;\r\nContent-Type: text/html; charset=\"UTF-8\";\r\n\r\n"
	msg += body
	return msg
}

// buildEmailBody 构建邮件正文（HTML格式）
func (en *EmailNotifier) buildEmailBody(alert Alert) string {
	tmpl := `
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; }
        .alert-box { 
            border-left: 4px solid {{.Color}}; 
            padding: 15px; 
            margin: 10px 0;
            background-color: #f9f9f9;
        }
        .alert-title { 
            font-size: 18px; 
            font-weight: bold; 
            color: {{.Color}};
            margin-bottom: 10px;
        }
        .alert-info { 
            margin: 5px 0; 
            color: #666;
        }
        .alert-message { 
            margin-top: 15px;
            padding: 10px;
            background-color: #fff;
            border: 1px solid #ddd;
            border-radius: 4px;
        }
        .metadata { 
            margin-top: 15px;
            padding: 10px;
            background-color: #f0f0f0;
            border-radius: 4px;
        }
    </style>
</head>
<body>
    <div class="alert-box">
        <div class="alert-title">{{.Emoji}} {{.Title}}</div>
        <div class="alert-info">
            <strong>级别:</strong> {{.Level}}<br>
            <strong>类型:</strong> {{.Type}}<br>
            <strong>时间:</strong> {{.Timestamp}}
        </div>
        <div class="alert-message">
            <strong>详细信息:</strong><br>
            {{.Message}}
        </div>
        {{if .Metadata}}
        <div class="metadata">
            <strong>附加信息:</strong><br>
            {{range $key, $value := .Metadata}}
            - {{$key}}: {{$value}}<br>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>
`
	
	data := struct {
		Title     string
		Level     string
		Type      string
		Message   string
		Timestamp string
		Color     string
		Emoji     string
		Metadata  map[string]interface{}
	}{
		Title:     alert.Title,
		Level:     alert.Level.String(),
		Type:      string(alert.Type),
		Message:   alert.Message,
		Timestamp: alert.Timestamp.Format("2006-01-02 15:04:05"),
		Color:     en.getLevelColor(alert.Level),
		Emoji:     en.getLevelEmoji(alert.Level),
		Metadata:  alert.Metadata,
	}
	
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		// 如果模板解析失败，返回简单文本
		return fmt.Sprintf(`
<h2>%s %s</h2>
<p><strong>级别:</strong> %s</p>
<p><strong>类型:</strong> %s</p>
<p><strong>时间:</strong> %s</p>
<p><strong>详细信息:</strong></p>
<pre>%s</pre>
`, 
			en.getLevelEmoji(alert.Level),
			alert.Title,
			alert.Level.String(),
			alert.Type,
			alert.Timestamp.Format("2006-01-02 15:04:05"),
			alert.Message,
		)
	}
	
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<pre>%s</pre>", alert.Message)
	}
	
	return buf.String()
}

// getLevelColor 获取级别对应的颜色
func (en *EmailNotifier) getLevelColor(level AlertLevel) string {
	switch level {
	case AlertLevelInfo:
		return "#3498db"
	case AlertLevelWarning:
		return "#f39c12"
	case AlertLevelCritical:
		return "#e74c3c"
	default:
		return "#95a5a6"
	}
}

// getLevelEmoji 获取级别对应的emoji
func (en *EmailNotifier) getLevelEmoji(level AlertLevel) string {
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

// SendSimpleText 发送简单文本邮件（用于测试）
func (en *EmailNotifier) SendSimpleText(subject, body string) error {
	message := en.buildMessage(subject, fmt.Sprintf("<pre>%s</pre>", body))
	
	addr := fmt.Sprintf("%s:%d", en.config.SMTPHost, en.config.SMTPPort)
	
	return smtp.SendMail(addr, en.auth, en.config.From, en.config.To, []byte(message))
}

