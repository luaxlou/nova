package main

import (
	"fmt"

	"github.com/luaxlou/nova/starter/glowconfig"
)

func main() {
	// Read configuration values directly using package-level functions
	// Config will be lazy-loaded on first access
	logLevel := glowconfig.GetString("log_level")
	maxConnections := glowconfig.GetInt("max_connections")
	mysqlDSN := glowconfig.GetString("mysql_dsn")
	redisAddr := glowconfig.GetString("redis_addr")

	// Print configuration
	fmt.Println("=== Glow Simple App ===")
	fmt.Printf("Log Level: %s\n", logLevel)
	fmt.Printf("Max Connections: %d\n", maxConnections)
	fmt.Printf("MySQL DSN: %s\n", mysqlDSN)
	fmt.Printf("Redis Addr: %s\n", redisAddr)

	// Check if running in Glow environment
	port := fmt.Sprintf("%s", glowconfig.Get("port"))
	if port != "" && port != "<nil>" {
		fmt.Printf("Port from config: %s\n", port)
	}

	fmt.Println("=== Application Started ===")
	fmt.Println("Hello, World!")
}
