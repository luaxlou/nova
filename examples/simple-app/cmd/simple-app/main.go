package main

import (
	"fmt"

	"github.com/luaxlou/nova/starter/config/novaconfig"
)

func main() {
	appName := novaconfig.GetString("app.name")
	maxConnections := novaconfig.GetInt("app.max_connections")

	fmt.Println("=== Nova Simple App ===")
	fmt.Printf("App Name: %s\n", appName)
	fmt.Printf("Max Connections: %d\n", maxConnections)

	fmt.Println("=== Application Started ===")
	fmt.Println("Hello, World!")
}
