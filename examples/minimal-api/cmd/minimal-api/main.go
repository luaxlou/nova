package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/glowhttp"
)

func main() {
	r := glowhttp.Router()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
	glowhttp.Init(8080)
	glowhttp.Run()
	select {}
}
