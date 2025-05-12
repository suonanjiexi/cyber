package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/suonanjiexi/cyber"
)

// CacheItem 缓存项
type CacheItem struct {
	Value      []byte
	Expiration time.Time
	Headers    map[string]string
	StatusCode int
}

// CacheStore 缓存存储接口
type CacheStore interface {
	Get(key string) (*CacheItem, bool)
	Set(key string, value *CacheItem, duration time.Duration)
	Delete(key string)
}

// MemoryStore 内存缓存存储
type MemoryStore struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
}

// NewMemoryStore 创建内存缓存存储
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		items: make(map[string]*CacheItem),
	}
	// 启动清理过期项的goroutine
	go store.startCleanup()
	return store
}

// Get 从缓存获取值
func (s *MemoryStore) Get(key string) (*CacheItem, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, found := s.items[key]
	if !found {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.Expiration) {
		s.mu.RUnlock()
		s.mu.Lock()
		delete(s.items, key)
		s.mu.Unlock()
		s.mu.RLock()
		return nil, false
	}

	return item, true
}

// Set 将值设置到缓存
func (s *MemoryStore) Set(key string, value *CacheItem, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	value.Expiration = time.Now().Add(duration)
	s.items[key] = value
}

// Delete 删除缓存项
func (s *MemoryStore) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

// startCleanup 启动清理过期项的定时任务
func (s *MemoryStore) startCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.cleanup()
	}
}

// cleanup 清理过期项
func (s *MemoryStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, item := range s.items {
		if now.After(item.Expiration) {
			delete(s.items, key)
		}
	}
}

// CacheConfig 缓存配置
type CacheConfig struct {
	TTL              time.Duration
	Store            CacheStore
	KeyPrefix        string
	CacheStatusCodes []int    // 要缓存的状态码列表，默认只缓存200
	IgnoreMethods    []string // 忽略的HTTP方法，默认忽略POST、PUT、DELETE、PATCH
	KeyGenerator     func(*cyber.Context) string
	CacheEmpty       bool     // 是否缓存空响应
	CacheHeaders     []string // 要保存在缓存中的响应头列表
}

// DefaultCacheConfig 默认缓存配置
var DefaultCacheConfig = CacheConfig{
	TTL:              5 * time.Minute,
	Store:            NewMemoryStore(),
	KeyPrefix:        "cyber-cache:",
	CacheStatusCodes: []int{200},
	IgnoreMethods:    []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
	KeyGenerator:     DefaultKeyGenerator,
	CacheEmpty:       false,
	CacheHeaders:     []string{"Content-Type", "Content-Length"},
}

// DefaultKeyGenerator 默认缓存键生成器
func DefaultKeyGenerator(c *cyber.Context) string {
	// 将URL、方法、查询参数和一些请求头合并为键
	path := c.Request.URL.Path
	method := c.Request.Method
	query := c.Request.URL.RawQuery

	// 排序查询参数以确保一致性
	if query != "" {
		params := strings.Split(query, "&")
		sort.Strings(params)
		query = strings.Join(params, "&")
	}

	// 包含一些请求头（如Accept、Accept-Encoding）
	var headers []string
	for _, h := range []string{"Accept", "Accept-Encoding"} {
		if v := c.Request.Header.Get(h); v != "" {
			headers = append(headers, h+":"+v)
		}
	}
	headerStr := ""
	if len(headers) > 0 {
		sort.Strings(headers)
		headerStr = "#" + strings.Join(headers, "|")
	}

	// 创建组合键
	key := fmt.Sprintf("%s-%s-%s%s", method, path, query, headerStr)

	// 使用SHA256哈希键以防止过长
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// ResponseCache 响应缓存中间件
func ResponseCache(next cyber.HandlerFunc) cyber.HandlerFunc {
	return ResponseCacheWithConfig(DefaultCacheConfig, next)
}

// ResponseCacheWithConfig 使用自定义配置的响应缓存中间件
func ResponseCacheWithConfig(config CacheConfig, next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		// 检查请求方法是否被忽略
		for _, method := range config.IgnoreMethods {
			if c.Request.Method == method {
				next(c)
				return
			}
		}

		// 生成缓存键
		key := config.KeyPrefix + config.KeyGenerator(c)

		// 尝试从缓存获取响应
		if item, found := config.Store.Get(key); found {
			// 从缓存设置响应头
			for name, value := range item.Headers {
				c.Writer.Header().Set(name, value)
			}

			// 设置缓存标记
			c.Writer.Header().Set("X-Cache", "HIT")

			// 写入状态码和响应体
			c.Status(item.StatusCode)
			c.Writer.Write(item.Value)
			return
		}

		// 创建响应记录器
		responseRecorder := &ResponseRecorder{
			ResponseWriter: c.Writer,
			Body:           &bytes.Buffer{},
			Headers:        make(http.Header),
		}

		// 替换原始ResponseWriter
		originalWriter := c.Writer
		c.Writer = responseRecorder

		// 处理请求
		next(c)

		// 恢复原始ResponseWriter
		c.Writer = originalWriter

		// 如果配置了缓存此状态码，则缓存响应
		shouldCache := false
		for _, code := range config.CacheStatusCodes {
			if responseRecorder.StatusCode == code {
				shouldCache = true
				break
			}
		}

		// 不缓存空响应（除非配置了要缓存）
		if responseRecorder.Body.Len() == 0 && !config.CacheEmpty {
			shouldCache = false
		}

		if shouldCache {
			// 保存要缓存的响应头
			headers := make(map[string]string)
			for _, headerName := range config.CacheHeaders {
				if values := responseRecorder.Header().Values(headerName); len(values) > 0 {
					headers[headerName] = values[0]
				}
			}

			// 创建缓存项
			item := &CacheItem{
				Value:      responseRecorder.Body.Bytes(),
				Headers:    headers,
				StatusCode: responseRecorder.StatusCode,
			}

			// 存储到缓存
			config.Store.Set(key, item, config.TTL)

			// 设置缓存标记
			responseRecorder.Header().Set("X-Cache", "MISS")
		}

		// 将记录的响应写入原始ResponseWriter
		for key, values := range responseRecorder.Header() {
			for _, value := range values {
				originalWriter.Header().Add(key, value)
			}
		}
		originalWriter.WriteHeader(responseRecorder.StatusCode)
		originalWriter.Write(responseRecorder.Body.Bytes())
	}
}

// ResponseRecorder 响应记录器，用于捕获响应内容
type ResponseRecorder struct {
	http.ResponseWriter
	StatusCode int
	Body       *bytes.Buffer
	Headers    http.Header
}

// WriteHeader 实现http.ResponseWriter的WriteHeader方法
func (r *ResponseRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write 实现http.ResponseWriter的Write方法
func (r *ResponseRecorder) Write(b []byte) (int, error) {
	r.Body.Write(b)
	return r.ResponseWriter.Write(b)
}

// Header 实现http.ResponseWriter的Header方法
func (r *ResponseRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}
