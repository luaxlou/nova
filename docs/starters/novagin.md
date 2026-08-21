# novagin

`starter/http/novagin` 提供基于 Gin 的 HTTP 服务启动适配。

## 配置来源

`novagin` 默认从配置读取端口。端口优先级如下：

1. `OP_APP_PORT` 环境变量
2. `http.port`
3. `PORT` 环境变量
4. 默认端口 `8080`

```yaml
http:
  port: 8080
```

## 最小用法

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/http/novagin"
)

func main() {
	r := novagin.Router()
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	novagin.Run()
	select {}
}
```

## 组合建议

- 最小 API 服务通常组合 `novaconfig` 与 `novagin`
- 需要数据库、缓存或 WebSocket 时，由应用层继续引入对应 starter
