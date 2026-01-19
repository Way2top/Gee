package gee

import (
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"
)

func Recovery() HandlerFunc {
	return func(c *Context) {
		defer func() {
			if err := recover(); err != nil {
				message := fmt.Sprintf("%s", err)
				log.Printf("%s\n\n", trace(message))
				c.Fail(http.StatusInternalServerError, "Internal Server Error")
			}
		}()
		c.Next()
	}
}

// trace 用于把函数调用路径展示出来，如果没有 trace，Recover() 仅仅只会简单的输出 panic 的原始消息，而 trace 可以将 panic 的消息加工处理，使得我们可以追根溯源找到问题所在
func trace(message string) string {
	var pcs [32]uintptr             // 用于记录函数调用顺序
	n := runtime.Callers(3, pcs[:]) // 从 goroutine 的调用栈中抓取“正在回退的函数路径”，跳过了前三层：Callers自己、trace 和 defer

	var str strings.Builder // 这个用于拼接字符串，至于为什么使用 Builder，是因为它是一个高性能拼接器，避免 str += ... 之类的操作
	str.WriteString(message + "\nTraceback:")
	for _, pc := range pcs[:n] { // 遍历 pcs，其中 pc 是某个函数正在执行的时候的 CPU 地址
		fn := runtime.FuncForPC(pc)                           // 把 CPU 地址翻译为函数对象，例如从 0x1054f8c1 变为 main.main.func2
		file, line := fn.FileLine(pc)                         // 从函数对象中查出源码的坐标，例如 /tmp/main.go:47
		str.WriteString(fmt.Sprintf("\n\t%s:%d", file, line)) // 把结果写入 str，最后返回
	}
	return str.String()
	// 返回的结果类似于：
	//Traceback:
	//	/tmp/main.go:47
	//	/tmp/context.go:41
	//	/tmp/recovery.go:37
	//	...
}
