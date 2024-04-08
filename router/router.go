package router

import "github.com/suonanjiexi/cyber"

type Router struct {
	app *cyber.App
}

func NewRouter(app *cyber.App) *Router {
	return &Router{app: app}
}
