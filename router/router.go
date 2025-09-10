package router

import (
	"claude2api/config"
	"claude2api/middleware"
	"claude2api/service"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// Apply middleware
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.AuthMiddleware())

	// Health check endpoint
	r.GET("/health", service.HealthCheckHandler)

	// WebSocket endpoint for real-time updates
	r.GET("/ws", service.WebSocketHandler)

	// Chat completions endpoint (OpenAI-compatible)
	r.POST("/v1/chat/completions", service.ChatCompletionsHandler)
	r.GET("/v1/models", service.ModelsHandler)

	// Admin routes for session management
	admin := r.Group("/admin")
	{
		admin.GET("/sessions", service.SessionsHealthHandler)
		admin.POST("/sessions", service.AddSessionHandler)
		admin.GET("/sessions/:key", service.SessionDetailHandler)
		admin.PUT("/sessions/:key", service.UpdateSessionHandler)
		admin.DELETE("/sessions/:key", service.DeleteSessionHandler)
		admin.POST("/sessions/:key/reset", service.ResetSessionHandler)
		admin.GET("/stats", service.SessionStatsHandler)
		admin.GET("/config", service.ConfigHandler)
		admin.PUT("/config", service.UpdateConfigHandler)
	}

	if config.ConfigInstance.EnableMirrorApi {
		r.POST(config.ConfigInstance.MirrorApiPrefix+"/v1/chat/completions", service.MirrorChatHandler)
		r.GET(config.ConfigInstance.MirrorApiPrefix+"/v1/models", service.ModelsHandler)
	}

	// HuggingFace compatible routes
	hfRouter := r.Group("/hf")
	{
		v1Router := hfRouter.Group("/v1")
		{
			v1Router.POST("/chat/completions", service.ChatCompletionsHandler)
			v1Router.GET("/models", service.ModelsHandler)
		}
	}
}
