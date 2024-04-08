package routers

import (
	"fmt"
	"github.com/suonanjiexi/cyber"
	"net/http"
)

func OrderRoutes(app *cyber.App) {
	//定义路由组
	order := app.Group("/order")
	order.Get("/detail", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("API Order")
		cyber.Success(w, r, http.StatusOK, "API Order")
	})
	order.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("API Order id ")
		cyber.Success(w, r, http.StatusOK, "API Order id ")
	})
}
