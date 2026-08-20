package main

import (
	"log"

	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/data"
	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/http"
	"github.com/luaxlou/nova/starter/novagin"
	"github.com/luaxlou/nova/starter/novagorm"
)

func main() {
	if err := novagorm.Init(); err != nil {
		log.Fatalf("gorm init failed: %v", err)
	}
	if err := initModels(); err != nil {
		log.Fatalf("model init failed: %v", err)
	}

	r := novagin.Router()

	userhttp.Routes(r)

	novagin.Init(8080)
	novagin.Run()

	log.Println("best-practice-service started")
	select {}
}

func initModels() error {
	db, err := novagorm.DB()
	if err != nil {
		return err
	}

	return db.AutoMigrate(&data.UserModel{})
}
