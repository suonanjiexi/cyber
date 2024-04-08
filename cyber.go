package cyber

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type HandlerFunc func(http.ResponseWriter, *http.Request)

type Middleware func(http.HandlerFunc) http.HandlerFunc

type App struct {
	Middlewares []Middleware
	Server      *http.Server
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
		Server: serverConfig,
	}
}

func (app *App) Use(middlewares ...Middleware) {
	app.Middlewares = append(app.Middlewares, middlewares...)
}

func applyMiddlewares(handler http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
func (app *App) HandleFunc(pattern string, handler http.HandlerFunc) {
	finalHandler := applyMiddlewares(handler, app.Middlewares)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic occurred in handler: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		finalHandler(w, r)
	})
	log.Printf("Route registered: %s", pattern)
}

func (app *App) Group(prefix string) *RouteGroup {
	return &RouteGroup{
		prefix: prefix,
		app:    app,
	}
}
func (rg *RouteGroup) HandleFunc(pattern string, handler http.HandlerFunc) {
	pattern = rg.joinPattern(pattern)
	rg.app.HandleFunc(pattern, handler)
}
func (rg *RouteGroup) joinPattern(pattern string) string {
	if !strings.HasPrefix(pattern, "/") {
		return strings.Join([]string{rg.prefix, "/", pattern}, "")
	}
	if rg.prefix != "/" && !strings.HasPrefix(rg.prefix, "/") {
		rg.prefix = "/" + rg.prefix
	}
	return strings.Join([]string{rg.prefix, pattern}, "")
}

func (app *App) Get(pattern string, handler http.HandlerFunc) {
	app.baseHttpHandler(http.MethodGet, pattern, handler)
}
func (app *App) Post(pattern string, handler http.HandlerFunc) {
	app.baseHttpHandler(http.MethodPost, pattern, handler)
}
func (app *App) Delete(pattern string, handler http.HandlerFunc) {
	app.baseHttpHandler(http.MethodDelete, pattern, handler)
}
func (app *App) Put(pattern string, handler http.HandlerFunc) {
	app.baseHttpHandler(http.MethodPut, pattern, handler)
}
func (app *App) Patch(pattern string, handler http.HandlerFunc) {
	app.baseHttpHandler(http.MethodPatch, pattern, handler)
}

func (rg *RouteGroup) Get(pattern string, handler http.HandlerFunc) {
	rg.baseHttpHandler(http.MethodGet, pattern, handler)
}

func (rg *RouteGroup) Post(pattern string, handler http.HandlerFunc) {
	rg.baseHttpHandler(http.MethodPost, pattern, handler)
}

func (rg *RouteGroup) Delete(pattern string, handler http.HandlerFunc) {
	rg.baseHttpHandler(http.MethodDelete, pattern, handler)
}

func (rg *RouteGroup) Put(pattern string, handler http.HandlerFunc) {
	rg.baseHttpHandler(http.MethodPut, pattern, handler)
}

func (rg *RouteGroup) Patch(pattern string, handler http.HandlerFunc) {
	rg.baseHttpHandler(http.MethodPatch, pattern, handler)
}

func isValidHTTPMethod(method string) bool {
	return strings.Contains("GET POST DELETE PUT PATCH", method)
}

func (rg *RouteGroup) baseHttpHandler(httpMethod string, pattern string, handler http.HandlerFunc) {
	if !isValidHTTPMethod(httpMethod) {
		log.Printf("Unsupported HTTP method: %s", httpMethod)
		return
	}
	rg.app.HandleFunc(httpMethod+" "+rg.prefix+pattern, handler)
}

func (app *App) baseHttpHandler(httpMethod string, pattern string, handler http.HandlerFunc) {
	if !isValidHTTPMethod(httpMethod) {
		log.Printf("Unsupported HTTP method: %s", httpMethod)
		return
	}
	app.HandleFunc(httpMethod+" "+pattern, handler)
}

// Run 添加启动成功的日志
func (app *App) Run() error {
	err := app.Server.ListenAndServe()
	if err != nil {
		log.Printf("Server failed to start: %v", err)
	}
	return err
}
func (app *App) Shutdown(ctx context.Context) error {
	err := app.Server.Shutdown(ctx)
	if err != nil {
		log.Printf("Server failed to Shutdown: %v", err)
	}
	return err
}
