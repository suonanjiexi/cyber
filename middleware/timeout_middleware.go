package middleware

import (
	"context"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

func TimeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxRetries := uint32(3)
		retry := uint32(0)
		timeout := 10 * time.Second
		for retry < maxRetries {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			r = r.WithContext(ctx)
			done := make(chan bool)
			go func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("Recovered in handler: %v", r)
					}
				}()
				next(w, r)
				done <- true
			}()
			select {
			case <-done:
				return
			case <-ctx.Done():
				retry = atomic.AddUint32(&retry, 1)
				if retry == maxRetries {
					log.Printf("Request timed out after maximum retries, last error: %v", ctx.Err())
					http.Error(w, "Request timed out after maximum retries", http.StatusGatewayTimeout)
					return
				}
				log.Printf("Request timed out, retrying (attempt %d)...", retry)
				timeout = doubleTimeout(timeout)
			}
		}
	}
}

func doubleTimeout(timeout time.Duration) time.Duration {
	const maxTimeout = 60 * time.Second
	if timeout < maxTimeout {
		return 2 * timeout
	}
	return maxTimeout
}
