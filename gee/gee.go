package gee

import (
	"log"
	"net/http"
	"strings"
)

// HandlerFunc 定义了Gee框架使用的请求处理函数类型
type HandlerFunc func(*Context)

// Engine 声明一个 Engine 结构体，这个结构体会实现 ServeHTTP 接口
type Engine struct {
	*RouterGroup
	router *router
	groups []*RouterGroup // 存储所有 groups
}

type RouterGroup struct {
	prefix      string
	middlewares []HandlerFunc // 支持中间件
	parent      *RouterGroup  // 支持嵌套
	engine      *Engine       // 所有 group 共享一个 Engine 实例
}

// New 创建 Engine 实例，初始化空的路由表
func New() *Engine {
	engine := &Engine{router: newRouter()}
	engine.RouterGroup = &RouterGroup{engine: engine}
	engine.groups = []*RouterGroup{engine.RouterGroup}
	return engine
	//return &Engine{router: newRouter()}
}

// Group 用于创建一个新的 RouterGroup
func (group *RouterGroup) Group(prefix string) *RouterGroup {
	engine := group.engine
	newGroup := &RouterGroup{
		prefix: group.prefix + prefix,
		parent: group,
		engine: engine,
	}
	engine.groups = append(engine.groups, newGroup)
	return newGroup
}

// 工具函数，后续 GET 和 POST 会使用这个函数给 engine 的路由表添加路由
func (group *RouterGroup) addRoute(method string, comp string, handler HandlerFunc) {
	pattern := group.prefix + comp
	log.Printf("Route %4s - %s", method, pattern)
	group.engine.router.addRoute(method, pattern, handler)
	//engine.router.addRoute(method, pattern, handler)
}

// GET 提供给用户注册 GET 请求的便捷方法
func (group *RouterGroup) GET(pattern string, handler HandlerFunc) {
	group.addRoute("GET", pattern, handler)
}

// POST 提供给用户注册 POST 请求的便捷方法
func (group *RouterGroup) POST(pattern string, handler HandlerFunc) {
	group.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

// Use 是用来将中间件添加到 group 的 middlewares[] 中的，group.middlewares[] 用于存储该路由分组下可能用到的中间件，而 Context.handlers[] 则存放实际需要用到的中间件，这是二者的区别所在
func (group *RouterGroup) Use(middlewares ...HandlerFunc) {
	group.middlewares = append(group.middlewares, middlewares...)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 这部分逻辑用于判断传入的路径所属的 group 包含哪些中间件
	var middlewares []HandlerFunc
	for _, group := range engine.groups {
		// 如果传入路径包含某个 group 的 prefix（前缀），说明传入的路径属于这个 group，那么这个 group 所包含的全部的中间件就都需要应用于该请求路径
		if strings.HasPrefix(req.URL.Path, group.prefix) {
			middlewares = append(middlewares, group.middlewares...)
		}
	}
	c := newContext(w, req)
	c.handlers = middlewares // 前面 for 循环找出了该请求路径下需要的全部中间件 middlewares，把这个加入到 Context.handlers，之后就可以根据 Context.handlers 来执行具体的中间件了
	engine.router.handle(c)
}
