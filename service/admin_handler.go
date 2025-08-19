package service

import (
	"claude2api/config"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionsHealthHandler 获取所有session的健康状态
func SessionsHealthHandler(c *gin.Context) {
	if !config.ConfigInstance.IsSessionManagerEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Session manager is not enabled",
		})
		return
	}

	sessionManager := config.ConfigInstance.GetSessionManager()
	sessions := sessionManager.GetSessionsHealth()

	c.JSON(http.StatusOK, gin.H{
		"sessions": sessions,
		"total":    len(sessions),
	})
}

// SessionStatsHandler 获取session统计信息
func SessionStatsHandler(c *gin.Context) {
	if !config.ConfigInstance.IsSessionManagerEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Session manager is not enabled",
		})
		return
	}

	sessionManager := config.ConfigInstance.GetSessionManager()
	stats := sessionManager.GetStats()

	c.JSON(http.StatusOK, stats)
}

// SessionDetailHandler 获取特定session的详细信息
func SessionDetailHandler(c *gin.Context) {
	sessionKey := c.Param("key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session key is required",
		})
		return
	}

	if !config.ConfigInstance.IsSessionManagerEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Session manager is not enabled",
		})
		return
	}

	sessionManager := config.ConfigInstance.GetSessionManager()
	sessions := sessionManager.GetSessionsHealth()

	for _, session := range sessions {
		if session.SessionKey == sessionKey {
			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Session not found",
	})
}

// ResetSessionHandler 重置特定session的状态
func ResetSessionHandler(c *gin.Context) {
	sessionKey := c.Param("key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session key is required",
		})
		return
	}

	if !config.ConfigInstance.IsSessionManagerEnabled() {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Session manager is not enabled",
		})
		return
	}

	// 这里可以实现重置session状态的逻辑
	// 目前只是返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"message": "Session reset request received",
		"session": sessionKey,
	})
}

// ConfigHandler 获取当前配置信息
func ConfigHandler(c *gin.Context) {
	configInfo := map[string]interface{}{
		"sessionManagerEnabled": config.ConfigInstance.IsSessionManagerEnabled(),
		"totalSessions":         len(config.ConfigInstance.Sessions),
		"retryCount":           config.ConfigInstance.RetryCount,
		"chatDelete":           config.ConfigInstance.ChatDelete,
		"maxChatHistoryLength": config.ConfigInstance.MaxChatHistoryLength,
	}

	if config.ConfigInstance.IsSessionManagerEnabled() {
		sm := config.ConfigInstance.GetSessionManager()
		configInfo["sessionManager"] = map[string]interface{}{
			"scheduleStrategy":     config.ConfigInstance.SessionManager.ScheduleStrategy,
			"maxRetryAttempts":     sm.GetMaxRetryAttempts(),
			"healthCheckInterval":  config.ConfigInstance.SessionManager.HealthCheckInterval,
			"minHealthScore":       config.ConfigInstance.SessionManager.MinHealthScore,
			"circuitBreakerEnabled": config.ConfigInstance.SessionManager.CircuitBreakerEnabled,
		}
	}

	c.JSON(http.StatusOK, configInfo)
}