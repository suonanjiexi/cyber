package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/suonanjiexi/cyber"
	"github.com/suonanjiexi/cyber/example/handler"
	"github.com/suonanjiexi/cyber/middleware"
)

func main() {
	// 创建应用
	app := cyber.NewApp(nil) // 使用默认配置

	// 添加全局中间件
	app.Use(
		middleware.Recovery,          // 首先添加恢复中间件，保证能够捕获其他中间件的崩溃
		middleware.Logger,            // 日志记录
		middleware.Cors,              // 跨域支持
		middleware.MetricsMiddleware, // 监控
		middleware.RateLimiter,       // 限流
	)

	// 创建用户处理器
	userHandler := handler.NewUserHandler()

	// 设置公开路由
	app.GET("/health", func(c *cyber.Context) {
		c.Success(200, map[string]string{"status": "ok"})
	})

	// 设置用户相关的路由，添加认证中间件
	userGroup := app.Group("/api/users")
	userGroup.Use(middleware.JWTAuth) // 为用户API添加认证中间件
	{
		// 获取所有用户
		userGroup.GET("", userHandler.GetAllUsers)
		// 创建用户
		userGroup.POST("", userHandler.CreateUser)
		// 获取单个用户
		userGroup.GET("/:id", userHandler.GetUser)
		// 更新用户
		userGroup.PUT("/:id", userHandler.UpdateUser)
		// 删除用户
		userGroup.DELETE("/:id", userHandler.DeleteUser)
	}

	// 启动服务器
	go func() {
		log.Println("服务器启动在 :8080")
		if err := app.Run(); err != nil {
			log.Fatalf("服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("正在关闭服务器...")

	// 关闭服务器
	if err := app.Shutdown(nil); err != nil {
		log.Fatalf("服务器关闭失败: %v", err)
	}
	log.Println("服务器已关闭")
}
