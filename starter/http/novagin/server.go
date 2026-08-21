package novagin

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/luaxlou/nova/starter/config/novaconfig"
)

var (
	engine      *gin.Engine
	once        sync.Once
	initialized bool
)

// Router returns the singleton Gin engine.
// It initializes the engine on the first call.
func Router() *gin.Engine {
	once.Do(func() {
		// Set Gin mode based on env or default to Release
		if os.Getenv("GIN_MODE") == "" {
			gin.SetMode(gin.ReleaseMode)
		}
		engine = gin.Default()

		// Add default middleware if needed
		// engine.Use(gin.Logger())
		// engine.Use(gin.Recovery())

		initialized = true
		log.Println("HTTP Starter (Gin) initialized.")
	})
	return engine
}

// Run starts the HTTP server on the configured port.
// It returns immediately, running the server in a goroutine.
func Run() {
	if !initialized {
		// Initialize if not already done (e.g. user didn't add any routes but just wants to start)
		Router()
	}

	runPort := ""

	// Priority: OP_APP_PORT > config http.port > PORT > 8080
	// 1. Check override from Nova Server (OP_APP_PORT)
	if p := os.Getenv("OP_APP_PORT"); p != "" {
		runPort = ":" + p
	}

	if runPort == "" {
		if p := novaconfig.GetInt("http.port"); p > 0 {
			runPort = ":" + strconv.Itoa(p)
		}
	}

	// 2. If no override and no configured port, check standard PORT env or default
	if runPort == "" {
		p := os.Getenv("PORT")
		if p == "" {
			p = "8080"
		}
		// If p is just a number, prefix with :
		if _, err := strconv.Atoi(p); err == nil {
			runPort = ":" + p
		} else {
			runPort = p
		}
	}

	srv := &http.Server{
		Addr:    runPort,
		Handler: engine,
	}

	go func() {
		log.Printf("Server starting on %s", runPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}
