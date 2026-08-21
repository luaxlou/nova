package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/cache/novaredis"
	"github.com/luaxlou/nova/starter/gorm/novagorm"
	"github.com/luaxlou/nova/starter/http/novagin"
)

func main() {
	r := novagin.Router()
	r.GET("/ready", func(c *gin.Context) {
		if _, err := novagorm.DB(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"gorm": err.Error()})
			return
		}
		if _, err := novaredis.Client(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	novagin.Run()
	log.Println("api-db-cache started")
	select {}
}
