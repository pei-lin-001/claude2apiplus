package middleware

import (
    "claude2api/config"
    "github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS headers with configurable allowed origins
func CORSMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        allowed := config.ConfigInstance.CORSAllowedOrigins
        origin := c.GetHeader("Origin")

        // 默认允许所有
        allowOrigin := ""
        if len(allowed) == 0 || (len(allowed) == 1 && allowed[0] == "*") {
            allowOrigin = "*"
        } else if origin != "" {
            for _, o := range allowed {
                if o == origin {
                    allowOrigin = origin
                    break
                }
            }
        }

        if allowOrigin != "" {
            c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
            c.Writer.Header().Set("Vary", "Origin")
        }
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization")
        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }
        c.Next()
    }
}
