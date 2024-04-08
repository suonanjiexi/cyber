package router

import (
	"net/http"
	"strings"
)

type Node struct {
	Children map[string]*Node
	Handler  http.HandlerFunc
	Pattern  string
	Wildcard bool
}

func NewNode() *Node {
	return &Node{
		Children: make(map[string]*Node),
	}
}

type Router struct {
	Root *Node
}

func NewRouter() *Router {
	return &Router{
		Root: NewNode(),
	}
}

func (r *Router) AddRoute(method, pattern string, handler http.HandlerFunc) {
	nodes := strings.Split(pattern, "/")

	// 从根节点开始构建Trie树
	current := r.Root
	for _, part := range nodes {
		if part == "*" {
			if _, ok := current.Children["*"]; !ok {
				current.Children["*"] = NewNode()
			}
			current = current.Children["*"]
			current.Wildcard = true
		} else if part == ":" {
			if _, ok := current.Children[":"]; !ok {
				current.Children[":"] = NewNode()
			}
			current = current.Children[":"]
			current.Wildcard = true
		} else {
			if _, ok := current.Children[part]; !ok {
				current.Children[part] = NewNode()
			}
			current = current.Children[part]
		}
	}
	current.Handler = handler
	current.Pattern = pattern
}

func (r *Router) HandleRequest(method, path string) (http.HandlerFunc, string) {
	nodes := strings.Split(path, "/")

	// 从根节点开始匹配路由
	current := r.Root
	for _, part := range nodes {
		if current.Wildcard {
			switch part {
			case "":
				// 如果是通配符节点，继续匹配下一个节点
				continue
			default:
				// 如果通配符节点有子节点，尝试匹配
				if child, ok := current.Children[part]; ok {
					current = child
				} else {
					// 没有匹配到路由
					return nil, ""
				}
			}
		} else if _, ok := current.Children[part]; !ok {
			// 没有匹配到路由
			return nil, ""
		}
		current = current.Children[part]
	}

	// 检查路由是否存在
	if current.Handler != nil {
		return current.Handler, current.Pattern
	}

	return nil, ""
}
