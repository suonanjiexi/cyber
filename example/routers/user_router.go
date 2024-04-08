package routers

import (
	"fmt"
	"github.com/suonanjiexi/cyber"
	"net/http"
)

func UserRoutes(app *cyber.App) {
	group := app.Group("/user")
	group.Get("/user", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("API User")
		cyber.Success(w, r, http.StatusOK, "API User")
	})
}
