package cyber

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
)

// Context 请求上下文
type Context struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	Params     map[string]string
	StatusCode int
	mutex      sync.RWMutex
	ctx        context.Context
	keys       map[string]interface{}
}

// 创建新的上下文
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		ctx:     r.Context(),
		keys:    make(map[string]interface{}),
	}
}

// GetParam 获取URL参数
func (c *Context) GetParam(key string) string {
	return c.Params[key]
}

// SetParam 设置URL参数
func (c *Context) SetParam(key, value string) {
	c.Params[key] = value
}

// Set 在上下文中存储键值对
func (c *Context) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.keys[key] = value
}

// Get 从上下文中获取键值对
func (c *Context) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	value, exists := c.keys[key]
	return value, exists
}

// MustGet 必须获取键值对，否则会panic
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key " + key + " does not exist")
}

// GetString 获取字符串类型的值
func (c *Context) GetString(key string) (string, error) {
	if val, ok := c.Get(key); ok {
		if str, ok := val.(string); ok {
			return str, nil
		}
		return "", errors.New("value is not a string")
	}
	return "", errors.New("key not found")
}

// GetInt 获取整数类型的值
func (c *Context) GetInt(key string) (int, error) {
	if val, ok := c.Get(key); ok {
		if i, ok := val.(int); ok {
			return i, nil
		}
		return 0, errors.New("value is not an integer")
	}
	return 0, errors.New("key not found")
}

// GetBool 获取布尔类型的值
func (c *Context) GetBool(key string) (bool, error) {
	if val, ok := c.Get(key); ok {
		if b, ok := val.(bool); ok {
			return b, nil
		}
		return false, errors.New("value is not a boolean")
	}
	return false, errors.New("key not found")
}

// Status 设置HTTP状态码
func (c *Context) Status(code int) {
	c.StatusCode = code
	c.Writer.WriteHeader(code)
}

// JSON 返回JSON格式的响应
func (c *Context) JSON(code int, obj interface{}) {
	c.Status(code)
	c.Writer.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
}

// String 返回字符串格式的响应
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Status(code)
	c.Writer.Header().Set("Content-Type", "text/plain")
	c.Writer.Write([]byte(format))
}

// HTML 返回HTML格式的响应
func (c *Context) HTML(code int, html string) {
	c.Status(code)
	c.Writer.Header().Set("Content-Type", "text/html")
	c.Writer.Write([]byte(html))
}

// Redirect 重定向
func (c *Context) Redirect(code int, location string) {
	c.Writer.Header().Set("Location", location)
	c.Status(code)
}

// Success 成功响应
func (c *Context) Success(code int, data interface{}) {
	c.JSON(code, data)
}

// Error 错误响应
func (c *Context) Error(code int, errCode string, message string) {
	c.JSON(code, map[string]interface{}{
		"code":    errCode,
		"message": message,
	})
}

// Abort 中止请求处理
func (c *Context) Abort() {
	panic("Abort")
}

// WithContext 设置新的上下文
func (c *Context) WithContext(ctx context.Context) *Context {
	c.ctx = ctx
	return c
}

// GetContext 获取上下文
func (c *Context) GetContext() context.Context {
	return c.ctx
}
