package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/luaxlou/nova/starter/novagin"
	"github.com/luaxlou/nova/starter/novawebsocket"
)

func main() {
	r := novagin.Router()
	r.GET("/ws", func(c *gin.Context) {
		novawebsocket.Handle(c, func(conn *websocket.Conn) {
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

	novagin.Init(8080)
	novagin.Run()
	log.Println("api-websocket started")
	select {}
}
