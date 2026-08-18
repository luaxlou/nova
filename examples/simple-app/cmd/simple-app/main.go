package main

import (
	"fmt"

	"github.com/luaxlou/nova/starter/novaconfig"
)

func main() {
	// Read configuration values directly using package-level functions
	// Config will be lazy-loaded on first access
	logLevel := novaconfig.GetString("log_level")
	maxConnections := novaconfig.GetInt("max_connections")
	mysqlDSN := novaconfig.GetString("mysql_dsn")
	redisAddr := novaconfig.GetString("redis_addr")

	// Print configuration
	fmt.Println("=== Nova Simple App ===")
	fmt.Printf("Log Level: %s\n", logLevel)
	fmt.Printf("Max Connections: %d\n", maxConnections)
	fmt.Printf("MySQL DSN: %s\n", mysqlDSN)
	fmt.Printf("Redis Addr: %s\n", redisAddr)

	// Check if running in Nova environment
	port := fmt.Sprintf("%s", novaconfig.Get("port"))
	if port != "" && port != "<nil>" {
		fmt.Printf("Port from config: %s\n", port)
	}

	fmt.Println("=== Application Started ===")
	fmt.Println("Hello, World!")
}
