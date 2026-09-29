package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"sync"
	"time"

	"blog/pkg/response"

	"github.com/gin-gonic/gin"
)

// rateBucket 滑动窗口限流的计数单元：记录窗口内每次请求的时间戳
type rateBucket struct {
	timestamps []time.Time // 请求时间戳列表
}

// RateLimiter 基于内存滑动窗口的限流器（单实例部署适用）
type RateLimiter struct {
	mu          sync.Mutex
	buckets     map[string]*rateBucket // 按 key（IP / 邮箱）区分
	limit       int                    // 窗口内允许的最大请求数
	window      time.Duration          // 窗口时长
	lastCleanup time.Time              // 上次过期桶清扫时间
}

// NewRateLimiter 创建限流器
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		buckets: make(map[string]*rateBucket),
		limit:   limit,
		window:  window,
	}
}

// Allow 判断 key 是否允许通过；允许则记录本次请求
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	// 定期清扫过期桶，避免大量伪造 key 导致内存无限增长
	r.maybeCleanup(now)

	bucket, ok := r.buckets[key]
	if !ok {
		// 首次请求，创建桶并记录
		bucket = &rateBucket{}
		r.buckets[key] = bucket
	}
	// 清理窗口外的过期时间戳
	cutoff := now.Add(-r.window)
	valid := bucket.timestamps[:0]
	for _, t := range bucket.timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	bucket.timestamps = valid

	if len(bucket.timestamps) >= r.limit {
		return false // 窗口内请求数已达上限
	}
	bucket.timestamps = append(bucket.timestamps, now)
	return true
}

// maybeCleanup 每个窗口周期清扫一次：剔除已无有效时间戳的桶，控制内存上限
func (r *RateLimiter) maybeCleanup(now time.Time) {
	if !r.lastCleanup.IsZero() && now.Sub(r.lastCleanup) < r.window {
		return
	}
	r.lastCleanup = now
	cutoff := now.Add(-r.window)
	for key, bucket := range r.buckets {
		valid := bucket.timestamps[:0]
		for _, t := range bucket.timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		bucket.timestamps = valid
		if len(valid) == 0 {
			delete(r.buckets, key)
		}
	}
}

// ClientIP 获取客户端真实 IP。
// 直接委托给 gin 的 c.ClientIP()：只有来自可信代理（router 层通过
// SetTrustedProxies 配置）的请求才会信任 X-Forwarded-For / X-Real-IP，
// 否则返回 TCP 直连地址，防止攻击者伪造请求头绕过限流。
func ClientIP(c *gin.Context) string {
	return c.ClientIP()
}

// RateLimit 限流中间件：按客户端 IP 限制单位时间内的请求次数，超出返回 42900
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)
	return func(c *gin.Context) {
		if !limiter.Allow(ClientIP(c)) {
			response.TooManyRequests(c, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginRateLimit 登录防爆破中间件：基于用户名+IP 限制尝试次数
// 失败 5 次后 15 分钟内禁止继续尝试（文档 B.2 建议）
func LoginRateLimit() gin.HandlerFunc {
	// 失败计数桶（key: username|ip）
	failLimiter := NewRateLimiter(5, 15*time.Minute)
	return func(c *gin.Context) {
		var body struct {
			Username string `json:"username"`
		}
		// 读取并恢复请求体，避免消费掉后续 handler 需要绑定的 JSON 数据
		data, err := io.ReadAll(c.Request.Body)
		if err == nil {
			c.Request.Body = io.NopCloser(bytes.NewReader(data))
			_ = json.Unmarshal(data, &body)
		}

		key := body.Username + "|" + ClientIP(c)
		if !failLimiter.Allow(key) {
			response.TooManyRequests(c, "登录尝试过于频繁，请 15 分钟后再试")
			c.Abort()
			return
		}
		c.Set("loginKey", key)
		c.Set("loginLimiter", failLimiter)
		c.Next()
	}
}

// ResetLoginFail 登录成功后清零失败计数（供 handler 调用）
func ResetLoginFail(c *gin.Context) {
	if key, ok := c.Get("loginKey"); ok {
		if v, ok := c.Get("loginLimiter"); ok {
			if limiter, ok := v.(*RateLimiter); ok {
				limiter.Delete(key.(string))
			}
		}
	}
}

// Delete 手动删除某个 key 的计数桶（用于登录成功后清零）
func (r *RateLimiter) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.buckets, key)
}
