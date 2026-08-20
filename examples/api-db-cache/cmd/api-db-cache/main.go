package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/novagin"
	"github.com/luaxlou/nova/starter/novamysql"
	"github.com/luaxlou/nova/starter/novaredis"
)

func main() {
	if err := novamysql.Init(); err != nil {
		log.Fatalf("mysql init failed: %v", err)
	}
	if err := novaredis.Init(); err != nil {
		log.Fatalf("redis init failed: %v", err)
	}

	r := novagin.Router()
	r.GET("/ready", func(c *gin.Context) {
		if _, err := novamysql.DB(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"mysql": err.Error()})
			return
		}
		if _, err := novaredis.Client(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"redis": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ready": true})
	})

	novagin.Init(8080)
	novagin.Run()
	log.Println("api-db-cache started")
	select {}
}
