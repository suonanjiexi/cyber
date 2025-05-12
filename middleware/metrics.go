package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/suonanjiexi/cyber"
)

// Metrics 指标统计结构体
type Metrics struct {
	TotalRequests     int64                      // 总请求数
	RequestsPerPath   map[string]int64           // 每个路径的请求数
	RequestsPerMethod map[string]int64           // 每个HTTP方法的请求数
	ResponseStatus    map[int]int64              // 每个状态码的请求数
	ResponseTimes     map[string][]time.Duration // 每个路径的响应时间
	ErrorCount        int64                      // 错误请求总数 (状态码 >= 400)
	ActiveRequests    int64                      // 当前活跃请求数
	MaxResponseTime   time.Duration              // 最长响应时间
	TotalResponseTime time.Duration              // 总响应时间
	Timestamp         time.Time                  // 开始统计的时间戳
	mu                sync.RWMutex               // 用于保护共享数据
}

// NewMetrics 创建指标统计实例
func NewMetrics() *Metrics {
	return &Metrics{
		RequestsPerPath:   make(map[string]int64),
		RequestsPerMethod: make(map[string]int64),
		ResponseStatus:    make(map[int]int64),
		ResponseTimes:     make(map[string][]time.Duration),
		Timestamp:         time.Now(),
	}
}

// 全局指标实例
var globalMetrics = NewMetrics()

// GetMetrics 获取全局指标
func GetMetrics() *Metrics {
	return globalMetrics
}

// Summary 获取格式化的指标摘要
func (m *Metrics) Summary() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 计算平均响应时间
	var avgResponseTime time.Duration
	if m.TotalRequests > 0 {
		avgResponseTime = time.Duration(m.TotalResponseTime.Nanoseconds() / m.TotalRequests)
	}

	// 计算路径级别的平均响应时间
	pathAvgResponseTime := make(map[string]string)
	for path, times := range m.ResponseTimes {
		if len(times) == 0 {
			continue
		}
		var total time.Duration
		for _, t := range times {
			total += t
		}
		avg := total / time.Duration(len(times))
		pathAvgResponseTime[path] = avg.String()
	}

	// 构建摘要
	return map[string]interface{}{
		"total_requests":      m.TotalRequests,
		"active_requests":     m.ActiveRequests,
		"error_count":         m.ErrorCount,
		"avg_response_time":   avgResponseTime.String(),
		"max_response_time":   m.MaxResponseTime.String(),
		"requests_per_method": m.RequestsPerMethod,
		"status_codes":        m.ResponseStatus,
		"path_avg_response":   pathAvgResponseTime,
		"uptime":              time.Since(m.Timestamp).String(),
	}
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalRequests = 0
	m.RequestsPerPath = make(map[string]int64)
	m.RequestsPerMethod = make(map[string]int64)
	m.ResponseStatus = make(map[int]int64)
	m.ResponseTimes = make(map[string][]time.Duration)
	m.ErrorCount = 0
	m.MaxResponseTime = 0
	m.TotalResponseTime = 0
	m.Timestamp = time.Now()
}

// RecordRequest 记录请求开始
func (m *Metrics) RecordRequest(path, method string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 增加总请求数
	atomic.AddInt64(&m.TotalRequests, 1)

	// 增加活跃请求数
	atomic.AddInt64(&m.ActiveRequests, 1)

	// 增加每个路径的请求数
	m.RequestsPerPath[path]++

	// 增加每个方法的请求数
	m.RequestsPerMethod[method]++
}

// RecordResponse 记录请求结束
func (m *Metrics) RecordResponse(path string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 减少活跃请求数
	atomic.AddInt64(&m.ActiveRequests, -1)

	// 增加每个状态码的请求数
	m.ResponseStatus[statusCode]++

	// 如果是错误响应，增加错误计数
	if statusCode >= 400 {
		atomic.AddInt64(&m.ErrorCount, 1)
	}

	// 记录响应时间
	m.ResponseTimes[path] = append(m.ResponseTimes[path], duration)

	// 更新总响应时间
	m.TotalResponseTime += duration

	// 更新最长响应时间
	if duration > m.MaxResponseTime {
		m.MaxResponseTime = duration
	}
}

// MetricsConfig 指标中间件配置
type MetricsConfig struct {
	SkipPaths []string // 不记录指标的路径
}

// DefaultMetricsConfig 默认指标中间件配置
var DefaultMetricsConfig = MetricsConfig{
	SkipPaths: []string{"/metrics", "/health", "/favicon.ico"},
}

// MetricsMiddleware 指标中间件
func MetricsMiddleware(next cyber.HandlerFunc) cyber.HandlerFunc {
	return MetricsMiddlewareWithConfig(DefaultMetricsConfig, next)
}

// MetricsMiddlewareWithConfig 使用自定义配置的指标中间件
func MetricsMiddlewareWithConfig(config MetricsConfig, next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		path := c.Request.URL.Path

		// 检查是否跳过此路径
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				next(c)
				return
			}
		}

		// 记录请求开始
		startTime := time.Now()
		method := c.Request.Method

		// 记录请求指标
		globalMetrics.RecordRequest(path, method)

		// 创建响应记录器以获取状态码
		responseRecorder := &StatusRecorder{
			ResponseWriter: c.Writer,
			StatusCode:     http.StatusOK, // 默认状态码
		}

		// 替换原始ResponseWriter
		c.Writer = responseRecorder

		// 处理请求
		next(c)

		// 计算响应时间并记录响应指标
		duration := time.Since(startTime)
		globalMetrics.RecordResponse(path, responseRecorder.StatusCode, duration)
	}
}

// StatusRecorder 记录HTTP响应状态码的ResponseWriter包装器
type StatusRecorder struct {
	http.ResponseWriter
	StatusCode int
}

// WriteHeader 记录状态码
func (r *StatusRecorder) WriteHeader(statusCode int) {
	r.StatusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write 实现http.ResponseWriter
func (r *StatusRecorder) Write(b []byte) (int, error) {
	return r.ResponseWriter.Write(b)
}

// Header 实现http.ResponseWriter
func (r *StatusRecorder) Header() http.Header {
	return r.ResponseWriter.Header()
}

// MetricsHandler 处理/metrics端点，返回当前指标数据
func MetricsHandler(c *cyber.Context) {
	metrics := GetMetrics()
	c.JSON(http.StatusOK, metrics.Summary())
}

// 注册指标处理器
func RegisterMetricsHandler(app *cyber.App) {
	app.GET("/metrics", MetricsHandler)
}

// 为指标服务提供简单的HTML视图
func MetricsViewHandler(c *cyber.Context) {
	metrics := GetMetrics()
	summary := metrics.Summary()

	// 将摘要转换为HTML表格
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Cyber 框架指标</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        h1 { color: #333; }
        table { border-collapse: collapse; width: 100%; margin-top: 20px; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
        tr:nth-child(even) { background-color: #f9f9f9; }
        .section { margin-top: 30px; }
        .refresh { margin-top: 20px; }
    </style>
</head>
<body>
    <h1>Cyber 框架指标</h1>
    <div class="refresh">
        <button onclick="location.reload()">刷新指标</button>
        <span>(上次更新: ` + time.Now().Format("2006-01-02 15:04:05") + `)</span>
    </div>
    
    <div class="section">
        <h2>概览</h2>
        <table>
            <tr><th>指标</th><th>值</th></tr>
            <tr><td>总请求数</td><td>` + fmt.Sprintf("%d", summary["total_requests"]) + `</td></tr>
            <tr><td>当前活跃请求数</td><td>` + fmt.Sprintf("%d", summary["active_requests"]) + `</td></tr>
            <tr><td>错误请求数</td><td>` + fmt.Sprintf("%d", summary["error_count"]) + `</td></tr>
            <tr><td>平均响应时间</td><td>` + summary["avg_response_time"].(string) + `</td></tr>
            <tr><td>最长响应时间</td><td>` + summary["max_response_time"].(string) + `</td></tr>
            <tr><td>运行时间</td><td>` + summary["uptime"].(string) + `</td></tr>
        </table>
    </div>
    
    <div class="section">
        <h2>按HTTP方法</h2>
        <table>
            <tr><th>方法</th><th>请求数</th></tr>
    `

	// 添加HTTP方法数据
	methodsData := summary["requests_per_method"].(map[string]int64)
	for method, count := range methodsData {
		html += "<tr><td>" + method + "</td><td>" + fmt.Sprintf("%d", count) + "</td></tr>"
	}

	html += `
        </table>
    </div>
    
    <div class="section">
        <h2>按状态码</h2>
        <table>
            <tr><th>状态码</th><th>请求数</th></tr>
    `

	// 添加状态码数据
	statusData := summary["status_codes"].(map[int]int64)
	for status, count := range statusData {
		html += "<tr><td>" + fmt.Sprintf("%d", status) + "</td><td>" + fmt.Sprintf("%d", count) + "</td></tr>"
	}

	html += `
        </table>
    </div>
    
    <div class="section">
        <h2>按路径的平均响应时间</h2>
        <table>
            <tr><th>路径</th><th>平均响应时间</th></tr>
    `

	// 添加路径响应时间数据
	pathData := summary["path_avg_response"].(map[string]string)
	for path, avgTime := range pathData {
		html += "<tr><td>" + path + "</td><td>" + avgTime + "</td></tr>"
	}

	html += `
        </table>
    </div>
</body>
</html>
    `

	c.HTML(http.StatusOK, html)
}

// 注册指标视图处理器
func RegisterMetricsViewHandler(app *cyber.App) {
	app.GET("/metrics/view", MetricsViewHandler)
}
