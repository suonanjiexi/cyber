package routers

import (
	"fmt"
	"github.com/suonanjiexi/cyber"
	"net/http"
)

func OrderRoutes(app *cyber.App) {
	//定义路由组
	group := app.Group("/order")
	group.HandleFunc("/detail", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "API Posts")
	})
}
