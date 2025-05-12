# Cyber - 一个高性能轻量级Go Web框架

Cyber是一个基于Go语言标准库实现的高性能轻量级Web框架，旨在提供简洁易用的API，同时保持极高的性能和极低的内存占用。

## 特性

* **轻量级**：基于"net/http"标准库，无过多依赖，极低的内存占用
* **高性能**：高效的前缀树路由匹配算法，中间件链优化
* **简洁API**：友好的接口设计，易于学习和使用
* **强大路由**：支持路由分组、参数化路由和通配符路由
* **灵活中间件**：支持全局、分组和路由级别的中间件
* **上下文管理**：请求上下文贯穿整个处理流程
* **数据验证**：内置请求数据验证功能
* **错误处理**：统一的错误处理机制
* **超时重试**：支持请求超时和自动重试
* **Go 1.22+**：利用最新Go版本特性

## 安装

```bash
go get -u github.com/suonanjiexi/cyber
```

## 快速开始

```go
package main

import (
	"github.com/suonanjiexi/cyber"
	"github.com/suonanjiexi/cyber/middleware"
	"net/http"
)

func main() {
	app := cyber.NewApp(nil)
	
	// 使用中间件
	app.Use(middleware.Recovery)
	app.Use(middleware.Logger)
	app.Use(middleware.Cors)
	
	// 注册路由
	app.GET("/", func(c *cyber.Context) {
		c.String(http.StatusOK, "Hello, Cyber!")
	})
	
	// 启动服务器
	app.Run()
}
```

## 路由系统

Cyber框架使用高效的前缀树（Trie树）实现路由匹配，支持静态路由、参数路由和通配符路由。

### 基本路由

```go
// 基本路由
app.GET("/users", GetUsers)
app.POST("/users", CreateUser)
app.PUT("/users/:id", UpdateUser)
app.DELETE("/users/:id", DeleteUser)
```

### 路由参数

```go
app.GET("/users/:id", func(c *cyber.Context) {
    id := c.GetParam("id")
    c.JSON(http.StatusOK, map[string]string{"id": id})
})
```

### 路由分组

```go
// 路由分组
api := app.Group("/api")
{
    v1 := api.Group("/v1")
    {
        v1.GET("/users", GetUsersV1)
        v1.POST("/users", CreateUserV1)
    }
    
    v2 := api.Group("/v2")
    {
        v2.GET("/users", GetUsersV2)
        v2.POST("/users", CreateUserV2)
    }
}
```

## 中间件系统

Cyber框架提供了一个强大且灵活的中间件系统，允许在HTTP请求处理过程中插入自定义逻辑。

### 中间件层级

框架支持三个层级的中间件：

1. **全局中间件**：应用于所有路由
2. **路由组中间件**：应用于特定路由组内的所有路由
3. **路由中间件**：仅应用于特定路由

### 编写中间件

```go
func LoggerMiddleware(next cyber.HandlerFunc) cyber.HandlerFunc {
    return func(c *cyber.Context) {
        // 请求处理前的逻辑
        start := time.Now()
        path := c.Request.URL.Path

        // 调用下一个中间件或最终处理函数
        next(c)

        // 请求处理后的逻辑
        log.Printf("[%s] %s %s %d %v", 
            c.Request.Method, 
            path, 
            c.Request.RemoteAddr, 
            c.StatusCode, 
            time.Since(start),
        )
    }
}
```

### 使用中间件

#### 全局中间件

```go
app := cyber.NewApp(nil)
app.Use(middleware.Recovery)
app.Use(middleware.Logger)
app.Use(middleware.Cors)
```

#### 路由组中间件

```go
api := app.Group("/api")
api.Use(AuthMiddleware)

// 所有/api/users路由都会应用AuthMiddleware
users := api.Group("/users")
users.GET("", listUsers)
users.POST("", createUser)
```

#### 路由特定中间件

```go
// 只有这个路由应用了AdminAuthMiddleware
app.GETWithMiddleware("/admin/settings", adminSettings, AdminAuthMiddleware)

// 路由组中也可以使用特定中间件
users.GETWithMiddleware("/:id", getUser, CacheMiddleware)
```

### 中间件执行顺序

多个中间件的执行顺序如下：

1. 全局中间件（按添加顺序）
2. 路由组中间件（按组嵌套顺序和添加顺序）
3. 路由特定中间件（按添加顺序）

### 内置中间件

框架提供了以下内置中间件：

1. **Recovery**：捕获panic并返回500错误响应
2. **Logger**：记录请求日志
3. **Cors**：处理跨域请求
4. **Timeout**：处理请求超时，支持重试

## 上下文和请求处理

Cyber框架使用Context对象贯穿整个请求处理流程，提供了丰富的方法来处理请求和生成响应。

```go
app.GET("/users/:id", func(c *cyber.Context) {
    id := c.GetParam("id")
    
    var user User
    // 绑定请求数据并验证
    if err := c.Bind(&user); err != nil {
        c.Error(http.StatusBadRequest, "INVALID_PARAMS", err.Error())
        return
    }
    
    // 返回JSON响应
    c.JSON(http.StatusOK, map[string]interface{}{
        "id": id,
        "name": user.Name,
    })
})
```

### 响应方法

- `c.JSON()`：返回JSON格式响应
- `c.String()`：返回文本响应
- `c.HTML()`：返回HTML响应
- `c.Redirect()`：执行重定向
- `c.Error()`：返回错误响应

### 上下文数据传递

```go
// 在中间件中设置数据
func AuthMiddleware(next cyber.HandlerFunc) cyber.HandlerFunc {
    return func(c *cyber.Context) {
        // 设置用户ID
        c.Set("userID", "123")
        next(c)
    }
}

// 在处理函数中获取数据
app.GET("/profile", func(c *cyber.Context) {
    userID, exists := c.Get("userID")
    if !exists {
        c.Error(http.StatusUnauthorized, "UNAUTHORIZED", "用户未认证")
        return
    }
    c.JSON(http.StatusOK, map[string]interface{}{
        "userID": userID,
    })
})
```

## 数据验证

Cyber框架内置了请求数据验证功能，支持多种验证规则。

```go
type User struct {
    Name  string `json:"name" valid:"required,min=2,max=50"`
    Email string `json:"email" valid:"required,email"`
    Age   int    `json:"age" valid:"min=18,max=150"`
}

app.POST("/users", func(c *cyber.Context) {
    var user User
    if err := c.Bind(&user); err != nil {
        c.Error(http.StatusBadRequest, "INVALID_PARAMS", err.Error())
        return
    }
    
    // 验证通过，继续处理
    c.JSON(http.StatusOK, user)
})
```

支持的验证规则：
- `required`：必填字段
- `min`：最小值/最小长度
- `max`：最大值/最大长度
- `email`：电子邮箱格式
- `pattern`：正则表达式匹配

## 完整示例

```go
package main

import (
	"context"
	"fmt"
	"github.com/suonanjiexi/cyber"
	"github.com/suonanjiexi/cyber/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// 创建应用实例
	config := &cyber.AppConfig{
		ServerPort:   "8080",
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
	}
	app := cyber.NewApp(config)
	
	// 全局中间件
	app.Use(middleware.Recovery)
	app.Use(middleware.Logger)
	app.Use(middleware.Cors)
	
	// 路由定义
	app.GET("/", func(c *cyber.Context) {
		c.String(http.StatusOK, "Hello, Cyber!")
	})
	
	api := app.Group("/api")
	{
		api.GET("/users", getUsers)
		api.POST("/users", createUser)
		api.GET("/users/:id", getUserByID)
	}
	
	// 启动服务器
	go func() {
		if err := app.Run(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	
	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := app.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
	}
	
	log.Println("Server exiting")
}

func getUsers(c *cyber.Context) {
	users := []map[string]interface{}{
		{"id": "1", "name": "用户1"},
		{"id": "2", "name": "用户2"},
	}
	c.JSON(http.StatusOK, users)
}

func createUser(c *cyber.Context) {
	type User struct {
		Name  string `json:"name" valid:"required,min=2,max=50"`
		Email string `json:"email" valid:"required,email"`
		Age   int    `json:"age" valid:"min=18,max=150"`
	}
	
	var user User
	if err := c.Bind(&user); err != nil {
		c.Error(http.StatusBadRequest, "INVALID_PARAMS", err.Error())
		return
	}
	
	c.JSON(http.StatusCreated, map[string]interface{}{
		"id":    "3",
		"name":  user.Name,
		"email": user.Email,
		"age":   user.Age,
	})
}

func getUserByID(c *cyber.Context) {
	id := c.GetParam("id")
	c.JSON(http.StatusOK, map[string]interface{}{
		"id":   id,
		"name": fmt.Sprintf("用户%s", id),
	})
}
```

## 性能优化

Cyber框架在设计上注重性能和内存使用，主要优化包括：

1. **高效的路由匹配**：采用前缀树实现，匹配时间复杂度为O(k)，k为URL路径长度
2. **中间件链优化**：预先组合中间件，减少运行时开销
3. **内存池化**：针对JSON编码器等对象使用对象池，减少GC压力
4. **零分配上下文**：尽可能减少上下文操作中的内存分配
5. **延迟加载**：某些组件采用延迟初始化，减少启动时开销

## 中间件最佳实践

1. **关注点分离**：每个中间件应专注于单一功能
2. **避免副作用**：中间件不应该改变请求数据，除非这是它的主要目的
3. **错误处理**：适当处理错误，并在必要时中止请求处理链
4. **考虑执行顺序**：某些中间件需要在其他中间件之前/之后执行
5. **使用上下文共享数据**：使用`c.Set()`和`c.Get()`在中间件和处理函数之间共享数据

## 贡献

欢迎提交PR或issue，一起完善这个框架！

## 许可证

MIT

## 版本

当前版本：v1.0.0
