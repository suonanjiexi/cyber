package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
