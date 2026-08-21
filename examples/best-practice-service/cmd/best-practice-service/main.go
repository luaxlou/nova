package main

import (
	"log"

	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/data"
	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/http"
	"github.com/luaxlou/nova/starter/gorm/novagorm"
	"github.com/luaxlou/nova/starter/http/novagin"
)

func main() {
	if err := initModels(); err != nil {
		log.Fatalf("model init failed: %v", err)
	}

	r := novagin.Router()

	userhttp.Routes(r)

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
