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
    
    // Public routes (no authentication)
    r.GET("/health", service.HealthCheckHandler)
    r.POST("/admin/login", service.AdminLoginHandler)
    r.POST("/admin/logout", service.AdminLogoutHandler)

    // API routes (API key authentication)
    api := r.Group("/v1")
    api.Use(middleware.AuthMiddleware())
    {
        api.POST("/chat/completions", service.ChatCompletionsHandler)
        api.GET("/models", service.ModelsHandler)
    }

    // WebSocket (require API key)
    r.GET("/ws", middleware.AuthMiddleware(), service.WebSocketHandler)

    // Admin routes (JWT authentication)
    admin := r.Group("/admin")
    admin.Use(middleware.AdminAuthMiddleware())
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
        admin.GET("/me", service.GetAdminInfoHandler)
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
