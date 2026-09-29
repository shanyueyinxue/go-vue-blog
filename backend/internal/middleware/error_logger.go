package middleware

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"blog/pkg/email"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// bodyLogWriter 包装 gin.ResponseWriter，缓冲响应体，
// 供 5xx 时记录后端返回的错误详情（gin.Logger 只有状态码，没有错误内容）
type bodyLogWriter struct {
	gin.ResponseWriter
	body bytes.Buffer
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// ErrorLog 记录服务器错误（HTTP >= 500）：方法、路径、客户端 IP、状态码与响应体片段。
// 开发环境仅记录日志；生产环境（notifier 非空且已启用）额外异步邮件通知站长，带冷却防风暴。
//
// 注意注册顺序：必须放在 gin.Recovery **之前**（更外层）。panic 时 Recovery 的 defer
// 会先写出 500 响应，之后 ErrorLog 的 c.Next() 之后代码上浮时才读得到 status=500；
// 若注册在 Recovery 之后，上浮时 ErrorLog 先执行，此时状态仍是 200，会漏记 panic。
func ErrorLog(logger *zap.Logger, notifier *ErrorNotifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		w := &bodyLogWriter{ResponseWriter: c.Writer}
		c.Writer = w
		c.Next()

		if status := c.Writer.Status(); status >= 500 {
			body := w.body.String()
			if len(body) > 512 {
				body = body[:512] + "..."
			}
			method := c.Request.Method
			path := c.Request.URL.Path
			ip := ClientIP(c)
			logger.Error("服务器错误",
				zap.Int("status", status),
				zap.String("method", method),
				zap.String("path", path),
				zap.String("ip", ip),
				zap.String("resp", body),
			)
			if notifier != nil && notifier.ShouldNotify() {
				notifier.NotifyAsync(method, path, ip, status, body)
				logger.Info("已向管理员发送邮件！")
			}
		}
	}
}

// defaultErrorNotifyCooldown 错误通知冷却时间：同一冷却窗口内的首个 5xx 触发邮件，
// 窗口内的后续 5xx 只记日志不再发邮件，防止故障风暴刷爆邮箱。
const defaultErrorNotifyCooldown = 5 * time.Minute

// ErrorNotifier 5xx 错误通知器：仅生产环境启用，异步邮件通知站长。
type ErrorNotifier struct {
	logger    *zap.Logger
	email     email.EmailService
	recipient string
	enabled   bool // 生产环境且站长邮箱已配置才启用
	cooldown  time.Duration

	mu         sync.Mutex
	lastSent   time.Time // 上次发送时间；零值表示从未发送
	suppressed int64     // 当前冷却窗口内被抑制（未发邮件）的错误数
}

// NewErrorNotifier 创建错误通知器。仅当 env 为 production 且邮件服务可用、站长邮箱
// 非空时启用；其余环境返回的 notifier 恒为禁用状态（仅记日志）。
func NewErrorNotifier(logger *zap.Logger, emailSvc email.EmailService, recipient, env string) *ErrorNotifier {
	recipient = strings.TrimSpace(recipient)
	return &ErrorNotifier{
		logger:    logger,
		email:     emailSvc,
		recipient: recipient,
		enabled:   env == "production" && emailSvc != nil && recipient != "",
		cooldown:  defaultErrorNotifyCooldown,
	}
}

// ShouldNotify 判断本次 5xx 是否应发送通知邮件（并发安全）：
// 距上次发送超过冷却时间则允许发送并更新 lastSent，否则抑制并计数。
func (n *ErrorNotifier) ShouldNotify() bool {
	if n == nil || !n.enabled {
		return false
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	now := time.Now()
	if n.lastSent.IsZero() || now.Sub(n.lastSent) >= n.cooldown {
		n.lastSent = now
		if n.suppressed > 0 {
			n.logger.Warn("错误通知冷却窗口结束", zap.Int64("suppressed", n.suppressed))
			n.suppressed = 0
		}
		return true
	}
	n.suppressed++
	return false
}

// NotifyAsync 异步发送错误通知邮件（defer/recover 保证不 panic，失败仅记 Warn 不影响主流程）
func (n *ErrorNotifier) NotifyAsync(method, path, ip string, status int, resp string) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				n.logger.Warn("发送错误通知邮件异常", zap.Any("recover", r))
			}
		}()
		msg := &email.Email{
			To:      []string{n.recipient},
			Subject: fmt.Sprintf("【博客告警】HTTP %d %s %s", status, method, path),
			Body: fmt.Sprintf(`<h3>服务器错误告警</h3>
<p>时间：%s</p>
<p>状态码：<strong>%d</strong></p>
<p>请求：<strong>%s %s</strong></p>
<p>客户端 IP：%s</p>
<p>响应内容：</p>
<pre style="background:#f6f6f6;padding:8px;border-radius:4px;">%s</pre>`,
				time.Now().Format("2006-01-02 15:04:05"), status, method, path, ip, resp),
		}
		if err := n.email.DialAndSend(msg); err != nil {
			n.logger.Warn("发送错误通知邮件失败", zap.Error(err), zap.String("to", n.recipient))
		}
	}()
}
