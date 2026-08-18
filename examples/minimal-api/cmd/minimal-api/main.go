package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/novahttp"
)

func main() {
	r := novahttp.Router()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, map[string]any{"ok": true})
	})
	novahttp.Init(8080)
	novahttp.Run()
	select {}
}
