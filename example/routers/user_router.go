package routers

import (
	"fmt"
	"github.com/suonanjiexi/cyber"
	"net/http"
)

func UserRoutes(app *cyber.App) {
	user := app.Group("/user")
	user.Get("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("API User")
		cyber.Success(w, r, http.StatusOK, "API User")
	})
	user.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		cyber.Success(w, r, http.StatusOK, "API Order id ")
	})
}
