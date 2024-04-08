package cyber

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LoggingMiddleware 为HTTP请求添加日志记录的中间件
func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 使用配置化的路径忽略列表，这里仅示例忽略了 favicon.ico
		ignorePaths := []string{"/favicon.ico"}
		requestPath := r.URL.Path
		isIgnored := false
		for _, path := range ignorePaths {
			if requestPath == path {
				isIgnored = true
				break
			}
		}
		// 如果请求不是被忽略的路径，则进行日志记录
		if !isIgnored {
			startTime := time.Now()
			defer logRequestDuration(startTime, r)
		}
		// 捕获并处理next函数可能引发的panic
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Recovered from panic: %v", err)
			}
		}()
		next(w, r)
	}
}

// logRequestDuration 记录请求的处理时间
func logRequestDuration(startTime time.Time, r *http.Request) {
	duration := time.Since(startTime)
	durationStr := formatDuration(duration)
	log.Printf("Duration: %s - Request: %s %s", durationStr, r.Method, r.URL.Path)
}

// formatDuration 根据持续时间返回格式化的时间字符串
func formatDuration(duration time.Duration) string {
	if duration.Minutes() >= 1 {
		return fmt.Sprintf("%.2f m", duration.Minutes())
	}
	if duration.Seconds() >= 1 {
		return fmt.Sprintf("%.2f s", duration.Seconds())
	}
	return fmt.Sprintf("%.2f ms", float64(duration.Nanoseconds())/float64(time.Millisecond))
}
