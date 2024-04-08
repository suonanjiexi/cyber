package cyber

import (
	"net/http"
	"strconv"
	"strings"
)

type CORSConfig struct {
	AllowOrigin   []string // 允许的来源列表
	AllowMethods  []string // 允许的请求方法列表
	AllowHeaders  []string // 允许的请求头部列表
	MaxAgeSeconds int      // Access-Control-Max-Age 的值，单位为秒
}

var defaultCORSConfig = CORSConfig{
	AllowOrigin:   []string{"*"},
	AllowMethods:  []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE"},
	AllowHeaders:  []string{"*"},
	MaxAgeSeconds: 3600,
}

func CorsMiddleware(next HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		headers := w.Header()
		headers.Add("Access-Control-Allow-Origin", strings.Join(defaultCORSConfig.AllowOrigin, ","))
		headers.Add("Access-Control-Allow-Methods", strings.Join(defaultCORSConfig.AllowMethods, ","))
		headers.Add("Access-Control-Allow-Headers", strings.Join(defaultCORSConfig.AllowHeaders, ","))
		if defaultCORSConfig.MaxAgeSeconds > 0 {
			headers.Add("Access-Control-Max-Age", strconv.Itoa(defaultCORSConfig.MaxAgeSeconds))
		}
		if r.Method == "OPTIONS" {
			return
		}
		next(w, r)
	}
}
