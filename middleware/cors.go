package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/suonanjiexi/cyber"
)

type CORSConfig struct {
	AllowOrigin      []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAgeSeconds    int
}

var defaultCORSConfig = CORSConfig{
	AllowOrigin:      []string{"*"},
	AllowMethods:     []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
	ExposeHeaders:    []string{},
	AllowCredentials: false,
	MaxAgeSeconds:    7200,
}

// Cors CORS中间件
func Cors(next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		headers := c.Writer.Header()

		// 设置允许的源
		headers.Set("Access-Control-Allow-Origin", strings.Join(defaultCORSConfig.AllowOrigin, ","))

		// 设置允许的方法
		headers.Set("Access-Control-Allow-Methods", strings.Join(defaultCORSConfig.AllowMethods, ","))

		// 设置允许的头部
		headers.Set("Access-Control-Allow-Headers", strings.Join(defaultCORSConfig.AllowHeaders, ","))

		// 设置暴露的头部
		if len(defaultCORSConfig.ExposeHeaders) > 0 {
			headers.Set("Access-Control-Expose-Headers", strings.Join(defaultCORSConfig.ExposeHeaders, ","))
		}

		// 设置是否允许凭证
		if defaultCORSConfig.AllowCredentials {
			headers.Set("Access-Control-Allow-Credentials", "true")
		}

		// 设置预检请求结果的缓存时间
		if defaultCORSConfig.MaxAgeSeconds > 0 {
			headers.Set("Access-Control-Max-Age", strconv.Itoa(defaultCORSConfig.MaxAgeSeconds))
		}

		// 对于预检请求，直接返回200响应
		if c.Request.Method == "OPTIONS" {
			c.Status(http.StatusOK)
			return
		}

		// 继续处理请求
		next(c)
	}
}
