package gee

import (
	"net/http"
	"strings"
)

type router struct {
	roots    map[string]*node
	handlers map[string]HandlerFunc
}

// roots key eg, roots['GET'] roots['POST']
// handlers key eg, handlers['GET-/p/:lang/doc'], handlers['POST-/p/book']
func newRouter() *router {
	return &router{
		// roots 的 key 是请求方式，值是具体的请求路径
		// eg: key = "GET", value = *node（Trie树的根节点，下面挂路径段）
		roots: make(map[string]*node),
		// handlers 的 key 是查找 handler 的唯一标识
		// eg: key = "GET-/user", value = HandlerFunc （真正执行的函数）
		handlers: make(map[string]HandlerFunc),
	}
}

// parsePattern 只做一件事：把路径字符串转化为 parts 数组（例如，"/p/:lang/doc" → ["p", ":lang", "doc"]）
func parsePattern(pattern string) []string {
	vs := strings.Split(pattern, "/")

	parts := make([]string, 0)
	for _, item := range vs {
		if item != "" {
			parts = append(parts, item)
			//parsePattern 中遇到 * 就 break，是因为 * 表示“匹配剩余所有路径”，逻辑上不允许再有更深的路由层级。
			// 举个例子，开发者设计了路由 /static/*filepath
			// 假设这里遇到了 * 并且不 break，会发生什么？比如设计了 /static/*filepath/abc
			// 如果不 break，会得到 ["static", "*filepath", "abc"]
			// 那么当一个真实的请求来了的时候，例如 /static/css/main.css
			// 你希望 *filepath = "css/main.css"
			// 但是如果 Trie 里还有一层 "abc"，那意味着你既让 *filepath 吃掉剩余所有路径，又要求后面必须再匹配 "abc"，这在语义上是自相矛盾的
			// 一句话总结就是，* 不是模糊匹配一个节点，而是终止 Trie 深度的兜底规则
			// : 和 * 是不同的，: 是吃一段，* 是吃剩下所有段
			// : 是占一个坑，* 是兜底首尾
			if item[0] == '*' {
				break
			}
		}
	}
	return parts
}

// 注册路由
func (r *router) addRoute(method string, pattern string, handler HandlerFunc) {
	parts := parsePattern(pattern)
	key := method + "-" + pattern
	_, ok := r.roots[method]
	if !ok {
		r.roots[method] = &node{}
	}
	r.roots[method].insert(pattern, parts, 0)
	r.handlers[key] = handler
}

// getRoute 输入请求方式和路径，返回匹配到的路由节点及 params（该方法在handler方法中被调用）
func (r *router) getRoute(method string, path string) (*node, map[string]string) {
	// 资源路径 pattern 被拆分为 parts（例如，"/p/:lang/doc" → ["p", ":lang", "doc"]）
	searchParts := parsePattern(path)
	params := make(map[string]string) // params 记录的就是动态路由匹配出来的参数，也就是 path 里 :name 或 *filepath 对应的实际值。
	root, ok := r.roots[method]       // 仅仅判断开发者有没有给该请求方法注册路由，例如有没有给 GET/POST 请求注册过路由

	// 如果连这个请求方法都没注册过，直接返回 nil
	if !ok {
		return nil, nil
	}

	// 如果注册过，接下来就开始正式进行动态路由匹配
	n := root.search(searchParts, 0)

	// n != nil 说明匹配到了路由
	if n != nil {
		// 匹配到了之后，把匹配到的 pattern 先解析为 parts
		parts := parsePattern(n.pattern)
		// 然后遍历解析后的 parts，去判断是不是动态参数(:) 或者通配符(*)
		for index, part := range parts {
			// 如果遇到路由以 : 开头，说明这部分是动态匹配的
			// 例如 :name，取掉冒号 part[1:] -> "name"，
			if part[0] == ':' {
				params[part[1:]] = searchParts[index]
				// params["name"] = "way2top"
				// params 的 key 是动态路由取掉冒号，value 是实际的路由
			}
			// 如果是 *，那么把剩下所有的路径拼成一个字符串写入 params
			if part[0] == '*' && len(part) > 1 {
				params[part[1:]] = strings.Join(searchParts[index:], "/")
				break
			}
		}
		return n, params
	}

	return nil, nil
}

// 根据 Context 查找 handler 并调用
func (r *router) handle(c *Context) {
	// n 是前缀树中匹配到的路由节点，例如请求 "/hello/way2top" 匹配到注册路由 "/hello/:name"
	// params 是具体的匹配值，例如 { "name": "way2top" }
	n, params := r.getRoute(c.Method, c.Path)
	if n != nil {
		c.Params = params
		key := c.Method + "-" + n.pattern
		//r.handlers[key](c)
		c.handlers = append(c.handlers, r.handlers[key])
	} else {
		c.handlers = append(c.handlers, func(c *Context) {
			c.String(http.StatusNotFound, "404 NOT FOUND: &s\n", c.Path)
		})
	}
	c.Next()
}
