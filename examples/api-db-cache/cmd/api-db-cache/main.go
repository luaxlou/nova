package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/glowhttp"
	"github.com/luaxlou/nova/starter/glowmysql"
	"github.com/luaxlou/nova/starter/glowredis"
)

func main() {
	glowmysql.Init("main")
	glowredis.Init()

	r := glowhttp.Router()
	r.GET("/ready", func(c *gin.Context) {
		if _, err := glowmysql.Gorm(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"mysql": err.Error()})
			return
		}
		if _, err := glowredis.Client(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	glowhttp.Init(8080)
	glowhttp.Run()
	log.Println("api-db-cache started")
	select {}
}
