# novawebsocket

`starter/realtime/novawebsocket` 提供基于 Gin 与 Gorilla WebSocket 的升级适配。

## 配置来源

`novawebsocket` 当前不读取 `config.yaml`。默认 Upgrader 使用以下参数：

```go
ReadBufferSize:  1024
WriteBufferSize: 1024
CheckOrigin:     func(*http.Request) bool { return true }
```

## 最小用法

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luaxlou/nova/starter/http/novagin"
	"github.com/luaxlou/nova/starter/realtime/novawebsocket"
)

func main() {
	r := novagin.Router()
	r.GET("/ws", func(c *gin.Context) {
		novawebsocket.Handle(c, func(conn *websocket.Conn) {
			_ = conn.WriteMessage(websocket.TextMessage, []byte("hello"))
		})
	})

	novagin.Run()
	select {}
}
```

## 约定

- WebSocket 路由通常由 `novagin.Router()` 提供
- 默认 `CheckOrigin` 是宽松策略，生产环境建议在应用层收紧来源校验
