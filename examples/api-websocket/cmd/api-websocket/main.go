package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luaxlou/nova/starter/glowhttp"
	"github.com/luaxlou/nova/starter/glowwebsocket"
)

func main() {
	r := glowhttp.Router()
	r.GET("/ws", func(c *gin.Context) {
		glowwebsocket.Handle(c, func(conn *websocket.Conn) {
			for {
				mt, msg, err := conn.ReadMessage()
				if err != nil {
					return
				}
				if err := conn.WriteMessage(mt, msg); err != nil {
					return
				}
			}
		})
	})

	glowhttp.Init(8080)
	glowhttp.Run()
	log.Println("api-websocket started")
	select {}
}
