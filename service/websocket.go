package service

import (
    "claude2api/config"
    "log"
    "net/http"
    "sync"
    "time"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var (
    upgrader = websocket.Upgrader{
        CheckOrigin: func(r *http.Request) bool {
            origins := config.ConfigInstance.CORSAllowedOrigins
            if len(origins) == 0 || (len(origins) == 1 && origins[0] == "*") {
                return true
            }
            reqOrigin := r.Header.Get("Origin")
            if reqOrigin == "" {
                return false
            }
            for _, o := range origins {
                if o == reqOrigin {
                    return true
                }
            }
            return false
        },
        ReadBufferSize:  1024,
        WriteBufferSize: 1024,
    }
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
	Timestamp time.Time `json:"timestamp"`
}

type WebSocketService struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WebSocketMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
	config     *config.Config
}

func NewWebSocketService(cfg *config.Config) *WebSocketService {
	return &WebSocketService{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan WebSocketMessage, 100),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		config:     cfg,
	}
}

func (ws *WebSocketService) Run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-ws.register:
			ws.mu.Lock()
			ws.clients[client] = true
			ws.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(ws.clients))

		case client := <-ws.unregister:
			ws.mu.Lock()
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				client.Close()
			}
			ws.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(ws.clients))

		case message := <-ws.broadcast:
			ws.mu.Lock()
			for client := range ws.clients {
				err := client.WriteJSON(message)
				if err != nil {
					log.Printf("WebSocket error: %v", err)
					client.Close()
					delete(ws.clients, client)
				}
			}
			ws.mu.Unlock()

		case <-ticker.C:
			// Broadcast stats update every 5 seconds
			if ws.config.IsSessionManagerEnabled() {
				ws.BroadcastStats()
			}
		}
	}
}

func (ws *WebSocketService) BroadcastStats() {
	if !ws.config.IsSessionManagerEnabled() {
		return
	}

	sessionManager := ws.config.GetSessionManager()
	stats := sessionManager.GetStats()

	message := WebSocketMessage{
		Type:      "stats_update",
		Data:      stats,
		Timestamp: time.Now(),
	}

	ws.broadcast <- message
}

func (ws *WebSocketService) BroadcastSessions() {
	if !ws.config.IsSessionManagerEnabled() {
		return
	}

	sessionManager := ws.config.GetSessionManager()
	sessions := sessionManager.GetSessionsHealth()

	message := WebSocketMessage{
		Type:      "sessions_update",
		Data:      sessions,
		Timestamp: time.Now(),
	}

	ws.broadcast <- message
}

func (ws *WebSocketService) BroadcastSessionChange(sessionKey string, action string) {
	message := WebSocketMessage{
		Type:      "session_change",
		Data: map[string]interface{}{
			"session_key": sessionKey,
			"action":      action,
		},
		Timestamp: time.Now(),
	}

	ws.broadcast <- message
}

func (ws *WebSocketService) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Register client
	ws.register <- conn

	// Send initial data
	if ws.config.IsSessionManagerEnabled() {
		sessionManager := ws.config.GetSessionManager()
		
		// Send initial stats
		stats := sessionManager.GetStats()
		initialStats := WebSocketMessage{
			Type:      "stats_update",
			Data:      stats,
			Timestamp: time.Now(),
		}
		conn.WriteJSON(initialStats)

		// Send initial sessions
		sessions := sessionManager.GetSessionsHealth()
		initialSessions := WebSocketMessage{
			Type:      "sessions_update",
			Data:      sessions,
			Timestamp: time.Now(),
		}
		conn.WriteJSON(initialSessions)
	}

	// Keep connection alive and listen for messages
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	// Unregister client
	ws.unregister <- conn
}

// Global WebSocket service instance
var WebSocketServiceInstance *WebSocketService

func InitWebSocketService(cfg *config.Config) {
	WebSocketServiceInstance = NewWebSocketService(cfg)
	go WebSocketServiceInstance.Run()
}

func WebSocketHandler(c *gin.Context) {
	if WebSocketServiceInstance == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "WebSocket service not initialized",
		})
		return
	}
	WebSocketServiceInstance.HandleWebSocket(c)
}
