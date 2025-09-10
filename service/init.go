package service

import (
	"claude2api/config"
)

// InitServices initializes all services
func InitServices(cfg *config.Config) {
	// Initialize WebSocket service
	InitializeWebSocketService(cfg)
}

// InitializeWebSocketService initializes the WebSocket service
func InitializeWebSocketService(cfg *config.Config) {
	// Use existing global instance initialization
	if WebSocketServiceInstance == nil {
		WebSocketServiceInstance = NewWebSocketService(cfg)
		go WebSocketServiceInstance.Run()
	}
}

// GetSessionManager returns the session manager from config
func GetSessionManager() interface{} {
	return config.ConfigInstance.GetSessionManager()
}