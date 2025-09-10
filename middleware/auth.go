package middleware

import (
	"claude2api/config"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 跳过登录路由
        if c.Request.URL.Path == "/admin/login" {
            c.Next()
            return
        }

		// 检查JWT token
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing authorization token"})
			c.Abort()
			return
		}

		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		
		// 解析JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(config.ConfigInstance.GetAdminSecret()), nil
		})

        if err != nil || !token.Valid {
            // 兼容旧版：若提供了 API Key，也允许访问（不建议，后续可移除）
            apiKey := c.GetHeader("Authorization")
            apiKey = strings.TrimPrefix(apiKey, "Bearer ")
            if apiKey != "" && apiKey == config.ConfigInstance.APIKey {
                c.Set("admin_user", "apikey")
                c.Next()
                return
            }
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

		// 设置用户信息到上下文
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("admin_user", claims["username"])
		}

		c.Next()
	}
}

// AuthMiddleware initializes the Claude client from the request header
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        if config.ConfigInstance.EnableMirrorApi && strings.HasPrefix(c.Request.URL.Path, config.ConfigInstance.MirrorApiPrefix) {
            c.Set("UseMirrorApi", true)
            c.Next()
            return
        }
        Key := c.GetHeader("Authorization")
        if Key != "" {
            Key = strings.TrimPrefix(Key, "Bearer ")
            if Key != config.ConfigInstance.APIKey {
                c.JSON(401, gin.H{
                    "error": "Invalid API key",
                })
                c.Abort()
                return
            }
            c.Next()
            return
        }
        // 兼容 WebSocket 或无法自定义头的场景：支持查询参数 token / api_key
        if c.Request.URL.Path == "/ws" {
            qp := c.Query("token")
            if qp == "" {
                qp = c.Query("api_key")
            }
            if qp != "" && qp == config.ConfigInstance.APIKey {
                c.Next()
                return
            }
        }
        c.JSON(401, gin.H{
            "error": "Missing or invalid Authorization header",
        })
        c.Abort()
    }
}

// GenerateAdminToken 生成管理员JWT token
func GenerateAdminToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.ConfigInstance.GetAdminSecret()))
}

// ValidateAdminCredentials 验证管理员凭据
func ValidateAdminCredentials(username, password string) bool {
	adminUser := config.ConfigInstance.GetAdminUser()
	adminPass := config.ConfigInstance.GetAdminPassword()
	
	return username == adminUser && password == adminPass
}
