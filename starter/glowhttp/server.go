package glowhttp

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	engine      *gin.Engine
	once        sync.Once
	initialized bool
	port        string
)

func Init(p int) {
	port = ":" + strconv.Itoa(p)
	// Initialize if explicitly called, though Run() will check too
	// Maybe we can load port here if we want to fail fast
}

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

	// Priority: OP_APP_PORT > Init() Port > PORT > 8080
	// 1. Check override from Glow Server (OP_APP_PORT)
	if p := os.Getenv("OP_APP_PORT"); p != "" {
		port = ":" + p
	}

	// 2. If no override and no Init() port, check standard PORT env or default
	if port == "" {
		p := os.Getenv("PORT")
		if p == "" {
			p = "8080"
		}
		// If p is just a number, prefix with :
		if _, err := strconv.Atoi(p); err == nil {
			port = ":" + p
		} else {
			port = p
		}
	}

	srv := &http.Server{
		Addr:    port,
		Handler: engine,
	}

	go func() {
		log.Printf("Server starting on %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
}
