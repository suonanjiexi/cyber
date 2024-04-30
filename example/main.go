package main

import (
	"context"
	"fmt"
	"github.com/suonanjiexi/cyber"
	"github.com/suonanjiexi/cyber/example/routers"
	"github.com/suonanjiexi/cyber/middleware"
	"log"
	"net/http"
)

func main() {
	app := cyber.NewApp(nil)
	// 使用中间件
	app.Use(middleware.Recovery)
	app.Use(middleware.Logger)
	app.Use(middleware.Cors)
	// 定义路由处理函数
	app.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hello, World!")
		cyber.Success(w, r, http.StatusOK, "Hello, World!")
	})
	routers.UserRoutes(app)
	routers.OrderRoutes(app)
	// 启动服务器
	if err := app.Run(); err != nil {
		log.Printf("Server error: %v", err)
	}
	// 关闭服务器
	if err := app.Shutdown(context.Background()); err != nil {
		log.Printf("Failed to shutdown server: %v", err)
	}
}
