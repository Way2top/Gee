package gee

import (
	"net/http"
)

// HandlerFunc 定义了Gee框架使用的请求处理函数类型
type HandlerFunc func(*Context)

// Engine 声明一个 Engine 结构体，这个结构体会实现 ServeHTTP 接口
type Engine struct {
	router *router
}

// New 创建 Engine 实例，初始化空的路由表
func New() *Engine {
	return &Engine{router: newRouter()}
}

// 工具函数，后续 GET 和 POST 会使用这个函数给 engine 的路由表添加路由
func (engine *Engine) addRoute(method string, pattern string, handler HandlerFunc) {
	engine.router.addRoute(method, pattern, handler)
}

// GET 提供给用户注册 GET 请求的便捷方法
func (engine *Engine) GET(pattern string, handler HandlerFunc) {
	engine.addRoute("GET", pattern, handler)
}

// POST 提供给用户注册 POST 请求的便捷方法
func (engine *Engine) POST(pattern string, handler HandlerFunc) {
	engine.addRoute("POST", pattern, handler)
}

func (engine *Engine) Run(addr string) (err error) {
	return http.ListenAndServe(addr, engine)
}

func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := newContext(w, req)
	engine.router.handle(c)
}
