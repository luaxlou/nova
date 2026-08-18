package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/novahttp"
	"github.com/luaxlou/nova/starter/novamysql"
	"github.com/luaxlou/nova/starter/novaredis"
)

func main() {
	novamysql.Init("main")
	novaredis.Init()

	r := novahttp.Router()
	r.GET("/ready", func(c *gin.Context) {
		if _, err := novamysql.Gorm(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"mysql": err.Error()})
			return
		}
		if _, err := novaredis.Client(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	novahttp.Init(8080)
	novahttp.Run()
	log.Println("api-db-cache started")
	select {}
}
