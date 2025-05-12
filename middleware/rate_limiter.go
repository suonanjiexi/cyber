package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/suonanjiexi/cyber"
)

// TokenBucket 令牌桶实现
type TokenBucket struct {
	tokens     float64
	capacity   float64
	rate       float64 // 每秒添加的令牌数
	lastRefill time.Time
	mu         sync.Mutex
}

// NewTokenBucket 创建新的令牌桶
func NewTokenBucket(capacity float64, rate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		rate:       rate,
		lastRefill: time.Now(),
	}
}

// Take 尝试从桶中获取token
func (tb *TokenBucket) Take() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// 计算从上次更新到现在需要添加的令牌数
	now := time.Now()
	elapsedSeconds := now.Sub(tb.lastRefill).Seconds()
	tb.lastRefill = now

	// 添加令牌
	tb.tokens += elapsedSeconds * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	// 检查是否有足够的令牌
	if tb.tokens < 1 {
		return false
	}

	// 消费一个令牌
	tb.tokens--
	return true
}

// RateLimiterConfig 限速器配置
type RateLimiterConfig struct {
	Rate     float64       // 每秒允许的请求数
	Capacity float64       // 令牌桶容量
	Timeout  time.Duration // 超过速率限制时的响应延迟
}

var defaultRateLimiterConfig = RateLimiterConfig{
	Rate:     10.0, // 每秒10个请求
	Capacity: 20.0, // 最多积累20个令牌
	Timeout:  0,    // 默认不延迟
}

// rateLimiterStore 保存IP地址到令牌桶的映射
var rateLimiterStore = struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
}{
	buckets: make(map[string]*TokenBucket),
}

// RateLimiter 速率限制中间件，使用IP地址作为标识
func RateLimiter(next cyber.HandlerFunc) cyber.HandlerFunc {
	return RateLimiterWithConfig(defaultRateLimiterConfig, next)
}

// RateLimiterWithConfig 使用自定义配置的速率限制中间件
func RateLimiterWithConfig(config RateLimiterConfig, next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		// 获取客户端IP
		ip := getClientIP(c.Request)

		// 获取或创建令牌桶
		rateLimiterStore.mu.RLock()
		bucket, exists := rateLimiterStore.buckets[ip]
		rateLimiterStore.mu.RUnlock()

		if !exists {
			bucket = NewTokenBucket(config.Capacity, config.Rate)
			rateLimiterStore.mu.Lock()
			rateLimiterStore.buckets[ip] = bucket
			rateLimiterStore.mu.Unlock()
		}

		// 尝试获取令牌
		if !bucket.Take() {
			// 如果配置了超时，则等待
			if config.Timeout > 0 {
				time.Sleep(config.Timeout)
				// 再次尝试获取令牌
				if !bucket.Take() {
					c.Error(http.StatusTooManyRequests, "RATE_LIMITED", "请求频率超过限制，请稍后再试")
					return
				}
			} else {
				c.Error(http.StatusTooManyRequests, "RATE_LIMITED", "请求频率超过限制，请稍后再试")
				return
			}
		}

		// 继续处理请求
		next(c)
	}
}

// getClientIP 获取客户端IP地址
func getClientIP(r *http.Request) string {
	// 尝试从各种头部获取真实IP
	ipHeaders := []string{
		"X-Real-IP",
		"X-Forwarded-For",
		"CF-Connecting-IP", // Cloudflare
		"True-Client-IP",   // Akamai and Cloudflare
	}

	for _, header := range ipHeaders {
		ip := r.Header.Get(header)
		if ip != "" {
			return ip
		}
	}

	// 如果没有代理头部，使用RemoteAddr
	ip := r.RemoteAddr
	return ip
}
