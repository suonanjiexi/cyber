package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	app := NewApp(nil)
	// 使用中间件
	app.Use(RecoveryMiddleware)
	app.Use(LoggingMiddleware)
	app.Use(TimeoutMiddleware)

	// 定义路由处理函数
	app.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "Hello, World!")
	})
	// 添加一个超时中间件
	app.HandleFunc("/timeout", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Second) // Simulating a long operation
		fmt.Println(w, "Hello, World!")
	})

	// 定义路由组
	api := app.Group("/api")

	// 在路由组中定义路由处理函数
	api.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "API Users")
	})
	api.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(w, "API Posts")
	})

	// 启动服务器
	if err := app.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
