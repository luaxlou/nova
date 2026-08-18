package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/novagin"
)

func main() {
	r := novagin.Router()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
	novagin.Init(8080)
	novagin.Run()
	select {}
}
