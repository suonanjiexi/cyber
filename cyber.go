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
	for i := range middlewares {
		handler = middlewares[len(middlewares)-1-i](handler)
	}
	return handler
}

func (app *App) Handle(pattern string, method string, handler http.HandlerFunc) {
	if !isValidHTTPMethod(method) {
		log.Printf("Unsupported HTTP method: %s", method)
		return
	}
	finalHandler := applyMiddlewares(handler, app.Middlewares)
	http.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic occurred in handler: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()
		finalHandler(w, r)
	})
	log.Printf("Route registered: %s %s", method, pattern)
}

func (app *App) Group(prefix string) *RouteGroup {
	return &RouteGroup{prefix: prefix, app: app}
}

func (rg *RouteGroup) Handle(pattern string, method string, handler http.HandlerFunc) {
	fullPattern := rg.joinPattern(pattern)
	rg.app.Handle(fullPattern, method, handler)
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
	allowedMethods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}
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
	return app.Server.ListenAndServe()
}

func (app *App) Shutdown(ctx context.Context) error {
	log.Printf("Shutting down server on %s", app.Server.Addr)
	return app.Server.Shutdown(ctx)
}
