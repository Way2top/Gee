package gee

import (
	"log"
	"net/http"
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

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
