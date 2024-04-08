package cyber

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Middleware func(HandlerFunc) HandlerFunc

type App struct {
	middlewares []Middleware
	server      *http.Server
}

type RouteGroup struct {
	prefix string
	app    *App
}

func NewApp(config *AppConfig) *App {
	return &App{
		server: &http.Server{
			Addr:         getValueWithDefault(config.ServerPort, defaultServerPort),
			ReadTimeout:  getDurationWithDefault(config.ReadTimeout, defaultReadTimeout),
			WriteTimeout: getDurationWithDefault(config.ReadTimeout, defaultWriteTimeout),
		},
	}
}

func (a *App) Use(middleware Middleware) {
	a.middlewares = append(a.middlewares, middleware)
}

func (a *App) HandleFunc(pattern string, handler HandlerFunc) {
	finalHandler := wrapHandler(handler, a.middlewares)
	http.HandleFunc(pattern, finalHandler)
	log.Printf("Route registered: %s", pattern)
}

func wrapHandler(handler HandlerFunc, middlewares []Middleware) HandlerFunc {
	if len(middlewares) == 0 {
		return handler
	}
	return wrapHandler(middlewares[0](handler), middlewares[1:])
}

func (a *App) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		prefix: prefix,
		app:    a,
	}
}
func (rg *RouteGroup) HandleFunc(pattern string, handler HandlerFunc) {
	pattern = rg.joinPattern(pattern)
	rg.app.HandleFunc(pattern, handler)
}
func (rg *RouteGroup) joinPattern(pattern string) string {
	if strings.HasPrefix(pattern, "/") {
		return strings.Join([]string{rg.prefix, pattern}, "")
	}
	return strings.Join([]string{rg.prefix, "/", pattern}, "")
}
func (a *App) Run() error {
	err := a.server.ListenAndServe()
	if err != nil {
		log.Printf("Server failed to start: %v", err)
	}
	return err
}
func (a *App) Get(pattern string, handler HandlerFunc) {
	a.baseHttpHandler(http.MethodGet, pattern, handler)
}
func (a *App) Post(pattern string, handler HandlerFunc) {
	a.baseHttpHandler(http.MethodPost, pattern, handler)
}
func (a *App) Delete(pattern string, handler HandlerFunc) {
	a.baseHttpHandler(http.MethodDelete, pattern, handler)
}
func (a *App) Put(pattern string, handler HandlerFunc) {
	a.baseHttpHandler(http.MethodPut, pattern, handler)
}
func (a *App) Patch(pattern string, handler HandlerFunc) {
	a.baseHttpHandler(http.MethodPatch, pattern, handler)
}
func (a *App) baseHttpHandler(httpMethod string, pattern string, handler HandlerFunc) {
	a.HandleFunc(httpMethod+" "+pattern, handler)
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.server.Shutdown(ctx)
}
func SafeInputMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		next(w, r)
	}
}
func RecoveryMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		next(w, r)
	}
}

// LoggingMiddleware 在请求处理完成后记录日志。
// 建议根据日志级别和业务需求调整日志记录策略，避免对性能的影响。
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

// TimeoutMiddleware 为请求设置超时。
// 建议根据具体业务场景评估是否为每个请求设置超时，以及超时时间的设置。
func TimeoutMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxRetries := uint32(3)
		retry := uint32(0) // 使用atomic操作，保证线程安全
		timeout := 10 * time.Second
		for retry < maxRetries {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel() // 确保在函数结束时调用cancel，避免资源泄露
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
