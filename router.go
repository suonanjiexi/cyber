package cyber

import (
	"strings"
)

// 路由节点
type node struct {
	children    map[string]*node
	handler     HandlerFunc
	pattern     string
	isParameter bool   // 是否是参数节点，例如 :id
	isWildcard  bool   // 是否是通配符节点，例如 *
	paramName   string // 参数名称，例如 :id 中的 id
}

func newNode() *node {
	return &node{
		children: make(map[string]*node),
	}
}

// 路由树，每个HTTP方法都有一个根节点
type trie struct {
	root *node
}

// 标准路由器实现
type StandardRouter struct {
	trees map[string]*trie // 每个HTTP方法对应一个前缀树
}

// 创建新的路由器
func NewRouter() Router {
	return &StandardRouter{
		trees: make(map[string]*trie),
	}
}

// 添加路由
func (r *StandardRouter) AddRoute(method, pattern string, handler HandlerFunc) {
	// 确保每个HTTP方法都有一个路由树
	if _, ok := r.trees[method]; !ok {
		r.trees[method] = &trie{root: newNode()}
	}

	// 标准化路径，确保以/开头，并去除末尾的/
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	if len(pattern) > 1 && strings.HasSuffix(pattern, "/") {
		pattern = pattern[:len(pattern)-1]
	}

	parts := strings.Split(pattern, "/")
	if parts[0] == "" {
		parts = parts[1:]
	}

	// 在对应的HTTP方法的路由树中插入路由
	current := r.trees[method].root
	for i, part := range parts {
		if part == "" {
			continue
		}

		isParameter := strings.HasPrefix(part, ":")
		isWildcard := part == "*"
		paramName := ""

		if isParameter {
			// 提取参数名
			paramName = strings.TrimPrefix(part, ":")
			part = ":" // 所有参数使用相同的节点
		}

		if _, ok := current.children[part]; !ok {
			current.children[part] = newNode()
			current.children[part].isParameter = isParameter
			current.children[part].isWildcard = isWildcard
			current.children[part].paramName = paramName
		}
		current = current.children[part]

		// 如果是最后一个部分，则设置handler
		if i == len(parts)-1 {
			current.handler = handler
			current.pattern = pattern
		}
	}
}

// 处理请求
func (r *StandardRouter) HandleRequest(c *Context) bool {
	// 获取请求方法和路径
	method := c.Request.Method
	path := c.Request.URL.Path

	// 确保路径以/开头
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	// 去除末尾的/
	if len(path) > 1 && strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}

	// 如果对应的HTTP方法没有路由树，则返回false
	tree, ok := r.trees[method]
	if !ok {
		return false
	}

	parts := strings.Split(path, "/")
	if parts[0] == "" {
		parts = parts[1:]
	}

	// 匹配路由
	params := make(map[string]string)
	if handler, matched := r.matchRoute(tree.root, parts, params); matched && handler != nil {
		// 将参数添加到上下文
		for k, v := range params {
			c.SetParam(k, v)
		}
		// 执行处理函数
		handler(c)
		return true
	}

	return false
}

// 匹配路由
func (r *StandardRouter) matchRoute(node *node, parts []string, params map[string]string) (HandlerFunc, bool) {
	if len(parts) == 0 {
		return node.handler, node.handler != nil
	}

	part := parts[0]
	rest := parts[1:]

	// 尝试精确匹配
	if child, ok := node.children[part]; ok {
		if handler, matched := r.matchRoute(child, rest, params); matched {
			return handler, true
		}
	}

	// 尝试参数匹配
	if child, ok := node.children[":"]; ok {
		// 保存参数值
		if child.paramName != "" {
			params[child.paramName] = part
		}
		if handler, matched := r.matchRoute(child, rest, params); matched {
			return handler, true
		}
	}

	// 尝试通配符匹配
	if child, ok := node.children["*"]; ok && child.handler != nil {
		return child.handler, true
	}

	return nil, false
}
