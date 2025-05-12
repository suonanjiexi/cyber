package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/suonanjiexi/cyber"
)

// Recovery 异常恢复中间件
func Recovery(next cyber.HandlerFunc) cyber.HandlerFunc {
	return func(c *cyber.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				log.Printf("Panic recovered: %v\nStack trace: %s", err, debug.Stack())
				// 响应500错误
				c.Error(http.StatusInternalServerError, "INTERNAL_ERROR", "Internal Server Error")
			}
		}()
		next(c)
	}
}
