package gee

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type H map[string]interface{}

type Context struct {
	Writer     http.ResponseWriter
	Req        *http.Request
	Path       string            // 请求资源路径
	Method     string            // 请求方式
	Params     map[string]string // 提供对路由参数的访问（router.go 中的 getRoute 返回的 params 就存储在这里）
	StatusCode int               // 状态码
	// 用于中间件
	handlers []HandlerFunc
	index    int // 用于控制中间件的执行顺序
}

func newContext(w http.ResponseWriter, req *http.Request) *Context {
	return &Context{
		Writer: w,
		Req:    req,
		Path:   req.URL.Path,
		Method: req.Method,
		index:  -1,
	}
}

// Next 用于控制中间件的执行顺序，调用后会将控制权交给下一个中间件
func (c *Context) Next() {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}

// PostForm 从 POST 表单中获取指定字段的值
func (c *Context) PostForm(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Query 从 URL 查询参数中获取指定字段的值。
func (c *Context) Query(key string) string {
	return c.Req.URL.Query().Get(key)
}

// Status 设置响应状态码并写入 HTTP Header
func (c *Context) Status(code int) {
	c.StatusCode = code        // 把状态码记录下来，方便以后打印日志、错误处理、中间件统计等
	c.Writer.WriteHeader(code) // 真正往 HTTP 响应码里写状态码
}

// SetHeader 设置响应头指定字段和值
func (c *Context) SetHeader(key string, value string) {
	c.Writer.Header().Set(key, value)
}

// String 以纯文本格式构造 HTTP 响应并写入客户端
func (c *Context) String(code int, format string, values ...interface{}) {
	c.SetHeader("Content-Type", "text/plain")
	c.Status(code)
	c.Writer.Write([]byte(fmt.Sprintf(format, values...)))
}

// JSON 将对象序列化为 JSON 格式big返回给客户端，同时设置 Content-Type
func (c *Context) JSON(code int, obj interface{}) {
	c.SetHeader("Content-Type", "application/json") // 设置 HTTP 响应头字段
	c.Status(code)
	encoder := json.NewEncoder(c.Writer)
	if err := encoder.Encode(obj); err != nil {
		http.Error(c.Writer, err.Error(), 500)
	}
}

// Data 直接写入原始字节数据到响应体，返回给客户端
func (c *Context) Data(code int, data []byte) {
	c.Status(code)
	c.Writer.Write(data)
}

// HTML 将 HTML 字符串作为响应返回，同时设置 Content-Type 为 text/html
func (c *Context) HTML(code int, html string) {
	c.SetHeader("Content-Type", "text/html")
	c.Status(code)
	c.Writer.Write([]byte(html))
}

// Param 可以取出 Context 中的 Params 的值，传入对应的 key 返回 Params[key]
func (c *Context) Param(key string) string {
	value, _ := c.Params[key]
	return value
}
