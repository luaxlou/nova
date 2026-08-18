package glowwebsocket

import (
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var (
	upgrader    *websocket.Upgrader
	initialized bool
	once        sync.Once
)

// DefaultUpgrader returns a singleton websocket.Upgrader with permissive defaults.
// (CheckOrigin returns true)
func DefaultUpgrader() *websocket.Upgrader {
	once.Do(func() {
		upgrader = &websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}
		initialized = true
		log.Println("WebSocket Starter initialized with default permissive upgrader.")
	})
	return upgrader
}

// HandlerFunc defines the signature for a websocket handler
type HandlerFunc func(*websocket.Conn)

// Handle is a helper to upgrade a Gin request to a WebSocket connection
// and pass it to the handler function.
func Handle(c *gin.Context, handler HandlerFunc) {
	upgrader := DefaultUpgrader()
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade websocket: %v", err)
		return
	}
	defer conn.Close()
	handler(conn)
}
