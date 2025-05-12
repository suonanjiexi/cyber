package middleware

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/suonanjiexi/cyber"
)

// 超时配置
type TimeoutConfig struct {
	Timeout    time.Duration // 超时时间
	MaxRetries uint32        // 最大重试次数
}

// 默认超时配置
var defaultTimeoutConfig = TimeoutConfig{
	Timeout:    10 * time.Second,
	MaxRetries: 3,
}

// Timeout 超时中间件
func Timeout(next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		var retry uint32
		timeout := defaultTimeoutConfig.Timeout

		for retry < defaultTimeoutConfig.MaxRetries {
			// 创建一个带超时的上下文
			ctx, cancel := context.WithTimeout(c.GetContext(), timeout)
			// 更新请求上下文
			c.WithContext(ctx)

			// 创建一个通道，用于标记请求是否完成
			done := make(chan bool, 1)

			// 在协程中执行请求处理
			go func() {
				defer func() {
					// 捕获可能的panic
					if r := recover(); r != nil {
						log.Printf("Recovered in TimeoutMiddleware: %v", r)
					}
					// 标记请求完成
					done <- true
				}()

				next(c)
			}()

			// 等待请求完成或超时
			select {
			case <-done:
				// 请求正常完成，取消上下文并返回
				cancel()
				return
			case <-ctx.Done():
				// 请求超时，重试
				retry = atomic.AddUint32(&retry, 1)
				if retry == defaultTimeoutConfig.MaxRetries {
					// 达到最大重试次数，返回超时响应
					log.Printf("Request timed out after %d retries, last error: %v", retry, ctx.Err())
					cancel()
					c.Error(http.StatusGatewayTimeout, "TIMEOUT", "Request timed out after maximum retries")
					return
				}

				log.Printf("Request timed out, retrying (attempt %d)...", retry)
				// 增加超时时间
				timeout = doubleTimeout(timeout)
				// 取消当前上下文
				cancel()
			}
		}
	}
}

// 翻倍超时时间，但不超过最大值
func doubleTimeout(timeout time.Duration) time.Duration {
	const maxTimeout = 60 * time.Second
	if timeout < maxTimeout {
		return 2 * timeout
	}
	return maxTimeout
}
