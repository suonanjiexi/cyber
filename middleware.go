package cyber

import (
	"log"
	"strings"
)

// MiddlewareChain 中间件链
type MiddlewareChain struct {
	middlewares []Middleware
}

// NewMiddlewareChain 创建新的中间件链
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Use 添加中间件到链中
func (mc *MiddlewareChain) Use(middlewares ...Middleware) *MiddlewareChain {
	mc.middlewares = append(mc.middlewares, middlewares...)
	return mc
}

// Clone 克隆中间件链
func (mc *MiddlewareChain) Clone() *MiddlewareChain {
	newChain := NewMiddlewareChain()
	newChain.middlewares = append(newChain.middlewares, mc.middlewares...)
	return newChain
}

// Apply 应用中间件链到处理函数
func (mc *MiddlewareChain) Apply(handler HandlerFunc) HandlerFunc {
	for i := range mc.middlewares {
		// 从右到左应用中间件，这样最先添加的中间件在最外层
		middleware := mc.middlewares[len(mc.middlewares)-1-i]
		handler = middleware(handler)
	}
	return handler
}

// Count 获取中间件数量
func (mc *MiddlewareChain) Count() int {
	return len(mc.middlewares)
}

// MiddlewareManager 中间件管理器
type MiddlewareManager struct {
	globalChain *MiddlewareChain                       // 全局中间件链
	groupChains map[string]*MiddlewareChain            // 路由组中间件链
	routeChains map[string]map[string]*MiddlewareChain // 路由中间件链 [方法][路径]
}

// NewMiddlewareManager 创建新的中间件管理器
func NewMiddlewareManager() *MiddlewareManager {
	return &MiddlewareManager{
		globalChain: NewMiddlewareChain(),
		groupChains: make(map[string]*MiddlewareChain),
		routeChains: make(map[string]map[string]*MiddlewareChain),
	}
}

// UseGlobal 添加全局中间件
func (mm *MiddlewareManager) UseGlobal(middlewares ...Middleware) {
	mm.globalChain.Use(middlewares...)
	log.Printf("Added %d global middleware(s), total: %d", len(middlewares), mm.globalChain.Count())
}

// UseGroup 添加路由组中间件
func (mm *MiddlewareManager) UseGroup(groupPrefix string, middlewares ...Middleware) {
	if _, exists := mm.groupChains[groupPrefix]; !exists {
		mm.groupChains[groupPrefix] = NewMiddlewareChain()
	}
	mm.groupChains[groupPrefix].Use(middlewares...)
	log.Printf("Added %d middleware(s) to group '%s', total: %d",
		len(middlewares), groupPrefix, mm.groupChains[groupPrefix].Count())
}

// UseRoute 添加路由中间件
func (mm *MiddlewareManager) UseRoute(method string, path string, middlewares ...Middleware) {
	if _, exists := mm.routeChains[method]; !exists {
		mm.routeChains[method] = make(map[string]*MiddlewareChain)
	}
	if _, exists := mm.routeChains[method][path]; !exists {
		mm.routeChains[method][path] = NewMiddlewareChain()
	}
	mm.routeChains[method][path].Use(middlewares...)
	log.Printf("Added %d middleware(s) to route '%s %s', total: %d",
		len(middlewares), method, path, mm.routeChains[method][path].Count())
}

// GetMiddlewareChain 获取指定路由的完整中间件链
func (mm *MiddlewareManager) GetMiddlewareChain(method string, path string) *MiddlewareChain {
	// 创建一个基于全局中间件的新链
	chain := mm.globalChain.Clone()

	// 添加匹配的组中间件（支持嵌套组）
	for prefix, groupChain := range mm.groupChains {
		if prefix == "/" || strings.HasPrefix(path, prefix) {
			// 路径匹配组前缀，添加组中间件
			chain.Use(groupChain.middlewares...)
		}
	}

	// 添加特定路由的中间件
	if methodChains, exists := mm.routeChains[method]; exists {
		if routeChain, exists := methodChains[path]; exists {
			chain.Use(routeChain.middlewares...)
		}
	}

	return chain
}

// ApplyMiddleware 应用所有中间件到处理函数
func (mm *MiddlewareManager) ApplyMiddleware(method string, path string, handler HandlerFunc) HandlerFunc {
	chain := mm.GetMiddlewareChain(method, path)
	return chain.Apply(handler)
}
