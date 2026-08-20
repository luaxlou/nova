package main

import (
	"log"

	"github.com/luaxlou/nova/examples/best-practice-service/internal/user/http"
	"github.com/luaxlou/nova/starter/novagin"
)

func main() {
	r := novagin.Router()

	userhttp.Routes(r)

	novagin.Init(8080)
	novagin.Run()

	log.Println("best-practice-service started")
	select {}
}
