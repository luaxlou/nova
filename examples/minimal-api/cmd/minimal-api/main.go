package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/http/novagin"
)

func main() {
	r := novagin.Router()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
	novagin.Run()
	select {}
}
