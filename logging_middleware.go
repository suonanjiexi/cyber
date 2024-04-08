package cyber

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func LoggingMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/favicon.ico" {
			startTime := time.Now()
			defer func() {
				duration := time.Since(startTime)
				var durationStr string
				switch {
				case duration.Minutes() >= 1:
					durationStr = fmt.Sprintf("%.2f m", duration.Minutes())
				case duration.Seconds() >= 1:
					durationStr = fmt.Sprintf("%.2f s", duration.Seconds())
				default:
					durationStr = fmt.Sprintf("%.2f ms", float64(duration.Nanoseconds())/float64(time.Millisecond))
				}
				log.Printf("Duration: %s - Request: %s %s", durationStr, r.Method, r.URL.Path)
			}()
		}
		next(w, r)
	}
}
