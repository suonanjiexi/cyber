### 一个轻量级go web框架
[详细文档](https://suonanjiexi.github.io/cyber/)
* go 1.22+ 
* 基于"net/http"标准库实现
* 支持中间件、路由分组、错误处理等
* 超时重试，幂等
* 低内存占用
* 路由树优化中...
```go
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


```
