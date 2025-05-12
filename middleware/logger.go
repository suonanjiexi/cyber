package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/suonanjiexi/cyber"
)

// Logger 日志中间件
func Logger(next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		// 请求开始时间
		startTime := time.Now()

		// 处理请求
		next(c)

		// 计算响应时间
		latency := time.Since(startTime)

		// 记录请求信息
		log.Printf(
			"[%s] %s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Request.RemoteAddr,
			c.StatusCode,
			latency,
		)
	}
}

func logRequestDuration(startTime time.Time, r *http.Request) {
	duration := time.Since(startTime)
	durationStr := formatDuration(duration)
	log.Printf("Duration: %s - Request: %s %s", durationStr, r.Method, r.URL.Path)
}

func formatDuration(duration time.Duration) string {
	if duration.Minutes() >= 1 {
		return fmt.Sprintf("%.2f m", duration.Minutes())
	}
	if duration.Seconds() >= 1 {
		return fmt.Sprintf("%.2f s", duration.Seconds())
	}
	return fmt.Sprintf("%.2f ms", float64(duration.Nanoseconds())/float64(time.Millisecond))
}
