package cyber

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// Middleware 中间件类型
type Middleware func(HandlerFunc) HandlerFunc

// Router 路由器接口
type Router interface {
	AddRoute(method, pattern string, handler HandlerFunc)
	HandleRequest(c *Context) bool
}

type App struct {
	Server            *http.Server
	Router            Router
	Config            *AppConfig
	MiddlewareManager *MiddlewareManager
}

type RouteGroup struct {
	prefix string
	app    *App
}

func NewApp(config *AppConfig) *App {
	if config == nil {
		config = &AppConfig{
			ServerPort:   defaultServerPort,
			ReadTimeout:  defaultReadTimeout,
			WriteTimeout: defaultWriteTimeout,
		}
	}

	serverConfig := &http.Server{
		Addr:         fmt.Sprintf(":%s", config.ServerPort),
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	}

	return &App{
		Server:            serverConfig,
		Router:            NewRouter(),
		Config:            config,
		MiddlewareManager: NewMiddlewareManager(),
	}
}

// Use 添加全局中间件
func (app *App) Use(middlewares ...Middleware) {
	app.MiddlewareManager.UseGlobal(middlewares...)
}

func (app *App) Handle(pattern string, method string, handler HandlerFunc) {
	if !isValidHTTPMethod(method) {
		log.Printf("Unsupported HTTP method: %s", method)
		return
	}

	// 应用中间件
	finalHandler := app.MiddlewareManager.ApplyMiddleware(method, pattern, handler)

	// 添加路由
	app.Router.AddRoute(method, pattern, finalHandler)
	log.Printf("Route registered: %s %s", method, pattern)
}

// HandleFunc is a shortcut for registering routes for any HTTP method
func (app *App) HandleFunc(pattern string, handler HandlerFunc) {
	// 默认注册为GET方法
	app.Handle(pattern, http.MethodGet, handler)
}

// 支持中间件的HTTP方法
func (app *App) GETWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	// 先注册路由特定的中间件
	app.MiddlewareManager.UseRoute(http.MethodGet, pattern, middlewares...)
	// 然后注册路由
	app.Handle(pattern, http.MethodGet, handler)
}

func (app *App) POSTWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodPost, pattern, middlewares...)
	app.Handle(pattern, http.MethodPost, handler)
}

func (app *App) PUTWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodPut, pattern, middlewares...)
	app.Handle(pattern, http.MethodPut, handler)
}

func (app *App) DELETEWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodDelete, pattern, middlewares...)
	app.Handle(pattern, http.MethodDelete, handler)
}

func (app *App) PATCHWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodPatch, pattern, middlewares...)
	app.Handle(pattern, http.MethodPatch, handler)
}

func (app *App) OPTIONSWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodOptions, pattern, middlewares...)
	app.Handle(pattern, http.MethodOptions, handler)
}

func (app *App) HEADWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	app.MiddlewareManager.UseRoute(http.MethodHead, pattern, middlewares...)
	app.Handle(pattern, http.MethodHead, handler)
}

func (app *App) Group(prefix string) *RouteGroup {
	return &RouteGroup{prefix: prefix, app: app}
}

func (rg *RouteGroup) Handle(pattern string, method string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, method, handler)
}

// 为RouteGroup添加支持中间件的HTTP方法
func (rg *RouteGroup) GETWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodGet, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodGet, handler)
}

func (rg *RouteGroup) POSTWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodPost, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodPost, handler)
}

func (rg *RouteGroup) PUTWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodPut, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodPut, handler)
}

func (rg *RouteGroup) DELETEWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodDelete, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodDelete, handler)
}

func (rg *RouteGroup) PATCHWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodPatch, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodPatch, handler)
}

func (rg *RouteGroup) OPTIONSWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodOptions, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodOptions, handler)
}

func (rg *RouteGroup) HEADWithMiddleware(pattern string, handler HandlerFunc, middlewares ...Middleware) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.MiddlewareManager.UseRoute(http.MethodHead, fullPattern, middlewares...)
	rg.app.Handle(fullPattern, http.MethodHead, handler)
}

func (rg *RouteGroup) Group(prefix string) *RouteGroup {
	fullPrefix := rg.joinPattern(prefix)
	return &RouteGroup{
		prefix: fullPrefix,
		app:    rg.app,
	}
}

// Use 添加路由组中间件
func (rg *RouteGroup) Use(middlewares ...Middleware) {
	groupPrefix := rg.prefix
	if !strings.HasPrefix(groupPrefix, "/") {
		groupPrefix = "/" + groupPrefix
	}
	rg.app.MiddlewareManager.UseGroup(groupPrefix, middlewares...)
}

func (rg *RouteGroup) joinPattern(pattern string) string {
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	if rg.prefix != "/" && !strings.HasPrefix(rg.prefix, "/") {
		rg.prefix = "/" + rg.prefix
	}
	return rg.prefix + pattern
}

func isValidHTTPMethod(method string) bool {
	allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions, http.MethodHead}
	for _, m := range allowedMethods {
		if m == method {
			return true
		}
	}
	return false
}

// Run logs the successful server start.
func (app *App) Run() error {
	log.Printf("Server starting on %s", app.Server.Addr)

	// 设置http.Handler
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		c := NewContext(w, r)
		// 处理请求
		if ok := app.Router.HandleRequest(c); !ok {
			// 没有找到匹配的路由
			http.NotFound(w, r)
		}
	})

	return app.Server.ListenAndServe()
}

func (app *App) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down server on %s", app.Server.Addr)
	return app.Server.Shutdown(ctx)
}

// HTTP方法便捷函数
func (app *App) GET(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodGet, handler)
}

func (app *App) POST(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodPost, handler)
}

func (app *App) PUT(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodPut, handler)
}

func (app *App) DELETE(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodDelete, handler)
}

func (app *App) PATCH(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodPatch, handler)
}

func (app *App) OPTIONS(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodOptions, handler)
}

func (app *App) HEAD(pattern string, handler HandlerFunc) {
	app.Handle(pattern, http.MethodHead, handler)
}

// 为RouteGroup添加HTTP方法便捷函数
func (rg *RouteGroup) GET(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodGet, handler)
}

func (rg *RouteGroup) POST(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodPost, handler)
}

func (rg *RouteGroup) PUT(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodPut, handler)
}

func (rg *RouteGroup) DELETE(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodDelete, handler)
}

func (rg *RouteGroup) PATCH(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodPatch, handler)
}

func (rg *RouteGroup) OPTIONS(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodOptions, handler)
}

func (rg *RouteGroup) HEAD(pattern string, handler HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, http.MethodHead, handler)
}
