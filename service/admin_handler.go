package service

import (
	"claude2api/config"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// formatDuration 格式化持续时间为秒级精度
func formatDuration(d time.Duration) string {
	seconds := int(d.Seconds())
	return fmt.Sprintf("%ds", seconds)
}

// SessionsHealthHandler 获取所有session的健康状态
func SessionsHealthHandler(c *gin.Context) {
	isEnabled := config.ConfigInstance.IsSessionManagerEnabled()
	
	if !isEnabled {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Session manager is not enabled",
			})
		return
	}

	sessionManager := config.ConfigInstance.GetSessionManager()
	sessions := sessionManager.GetSessionsHealth()

	c.JSON(http.StatusOK, sessions)
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

	// 添加系统运行时间统计
	systemStats := map[string]interface{}{
		"total_sessions":    len(config.ConfigInstance.Sessions),
		"active_sessions":   stats.SessionsActive,
		"cooling_sessions":  stats.SessionsCooling,
		"failed_sessions":   stats.SessionsFailed,
		"total_requests":    stats.TotalRequests,
		"successful_requests": stats.SuccessfulReqs,
		"failed_requests":  stats.FailedRequests,
		"average_latency":   stats.AverageLatency.Milliseconds(),
		"uptime":           formatDuration(time.Since(config.ConfigInstance.GetSessionManager().GetStartTime())),
		"last_reset":        stats.LastReset,
		"errors_by_type":    stats.ErrorsByType,
	}

	c.JSON(http.StatusOK, systemStats)
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

// AddSessionHandler 添加新的Session
func AddSessionHandler(c *gin.Context) {
	var req struct {
		SessionKey string `json:"sessionKey" binding:"required"`
		OrgID      string `json:"orgID"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	// 验证Session Key格式
	if !isValidSessionKey(req.SessionKey) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid session key format. Session key should start with 'sk-' or 'sk-ant-'",
		})
		return
	}

	// 检查Session是否已存在
	config.ConfigInstance.RwMutx.RLock()
	for _, session := range config.ConfigInstance.Sessions {
		if session.SessionKey == req.SessionKey {
			config.ConfigInstance.RwMutx.RUnlock()
			c.JSON(http.StatusConflict, gin.H{
				"error": "Session already exists",
			})
			return
		}
	}
	config.ConfigInstance.RwMutx.RUnlock()

	// 添加新Session到Config
	config.ConfigInstance.RwMutx.Lock()
	newSession := config.SessionInfo{
		SessionKey: req.SessionKey,
		OrgID:      req.OrgID,
	}
	config.ConfigInstance.Sessions = append(config.ConfigInstance.Sessions, newSession)
	config.ConfigInstance.RwMutx.Unlock()

	// 同时添加到SessionManager中（如果SessionManager已启用）
	if config.ConfigInstance.IsSessionManagerEnabled() {
		sessionManager := config.ConfigInstance.GetSessionManager()
		if sessionManager != nil {
			sessionManager.AddSession(newSession)
		}
	}

	// 广播WebSocket消息
	if WebSocketServiceInstance != nil {
		WebSocketServiceInstance.BroadcastSessionChange(req.SessionKey, "added")
		WebSocketServiceInstance.BroadcastSessions()
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Session added successfully",
		"session": newSession,
	})
}

// DeleteSessionHandler 删除Session
func DeleteSessionHandler(c *gin.Context) {
	sessionKey := c.Param("key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session key is required",
		})
		return
	}

	config.ConfigInstance.RwMutx.Lock()
	defer config.ConfigInstance.RwMutx.Unlock()

	// 查找并删除Session
	found := false
	for i, session := range config.ConfigInstance.Sessions {
		if session.SessionKey == sessionKey {
			config.ConfigInstance.Sessions = append(config.ConfigInstance.Sessions[:i], config.ConfigInstance.Sessions[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Session not found",
		})
		return
	}

	// 同时从SessionManager中移除session（如果SessionManager已启用）
	if config.ConfigInstance.IsSessionManagerEnabled() {
		sessionManager := config.ConfigInstance.GetSessionManager()
		if sessionManager != nil {
			sessionManager.RemoveSession(sessionKey)
		}
	}

	// 广播WebSocket消息
	if WebSocketServiceInstance != nil {
		WebSocketServiceInstance.BroadcastSessionChange(sessionKey, "deleted")
		WebSocketServiceInstance.BroadcastSessions()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session deleted successfully",
	})
}

// UpdateSessionHandler 更新Session信息
func UpdateSessionHandler(c *gin.Context) {
	sessionKey := c.Param("key")
	if sessionKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Session key is required",
		})
		return
	}

	var req struct {
		OrgID string `json:"orgID"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
		})
		return
	}

	config.ConfigInstance.RwMutx.Lock()
	defer config.ConfigInstance.RwMutx.Unlock()

	// 查找并更新Session
	found := false
	for i, session := range config.ConfigInstance.Sessions {
		if session.SessionKey == sessionKey {
			config.ConfigInstance.Sessions[i].OrgID = req.OrgID
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Session not found",
		})
		return
	}

	// 广播WebSocket消息
	if WebSocketServiceInstance != nil {
		WebSocketServiceInstance.BroadcastSessionChange(sessionKey, "updated")
		WebSocketServiceInstance.BroadcastSessions()
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Session updated successfully",
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

// UpdateConfigHandler 更新系统配置
func UpdateConfigHandler(c *gin.Context) {
	var req struct {
		SessionManagerEnabled    bool                   `json:"sessionManagerEnabled"`
		TotalSessions            int                    `json:"totalSessions"`
		RetryCount              int                    `json:"retryCount"`
		ChatDelete              bool                   `json:"chatDelete"`
		MaxChatHistoryLength    int                    `json:"maxChatHistoryLength"`
		NoRolePrefix            bool                   `json:"noRolePrefix"`
		PromptDisableArtifacts  bool                   `json:"promptDisableArtifacts"`
		EnableMirrorApi         bool                   `json:"enableMirrorApi"`
		MirrorApiPrefix         string                 `json:"mirrorApiPrefix"`
		SessionManager          *map[string]interface{} `json:"sessionManager"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 更新配置
	config.ConfigInstance.RwMutx.Lock()
	defer config.ConfigInstance.RwMutx.Unlock()

	config.ConfigInstance.SessionManager.Enabled = req.SessionManagerEnabled
	config.ConfigInstance.RetryCount = req.RetryCount
	config.ConfigInstance.ChatDelete = req.ChatDelete
	config.ConfigInstance.MaxChatHistoryLength = req.MaxChatHistoryLength
	config.ConfigInstance.NoRolePrefix = req.NoRolePrefix
	config.ConfigInstance.PromptDisableArtifacts = req.PromptDisableArtifacts
	config.ConfigInstance.EnableMirrorApi = req.EnableMirrorApi
	config.ConfigInstance.MirrorApiPrefix = req.MirrorApiPrefix

	// 更新SessionManager详细配置
	if req.SessionManager != nil {
		if strategy, ok := (*req.SessionManager)["scheduleStrategy"].(string); ok {
			config.ConfigInstance.SessionManager.ScheduleStrategy = strategy
		}
		if maxRetry, ok := (*req.SessionManager)["maxRetryAttempts"].(int); ok {
			config.ConfigInstance.SessionManager.MaxRetryAttempts = maxRetry
		}
		if healthInterval, ok := (*req.SessionManager)["healthCheckInterval"].(int); ok {
			config.ConfigInstance.SessionManager.HealthCheckInterval = time.Duration(healthInterval) * time.Second
		}
		if minHealth, ok := (*req.SessionManager)["minHealthScore"].(float64); ok {
			config.ConfigInstance.SessionManager.MinHealthScore = minHealth
		}
		if circuitBreaker, ok := (*req.SessionManager)["circuitBreakerEnabled"].(bool); ok {
			config.ConfigInstance.SessionManager.CircuitBreakerEnabled = circuitBreaker
		}
	}

	// 如果SessionManager状态改变，重新初始化
	if req.SessionManagerEnabled {
		if config.ConfigInstance.GetSessionManager() == nil {
			config.ConfigInstance.GetSessionManager()
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Configuration updated successfully"})
}

// isValidSessionKey 验证Session Key格式
func isValidSessionKey(sessionKey string) bool {
	if sessionKey == "" {
		return false
	}
	
	// 去除前后空格
	sessionKey = strings.TrimSpace(sessionKey)
	
	// 检查最小长度
	if len(sessionKey) < 10 {
		return false
	}
	
	// 检查是否以有效的prefix开头
	validPrefixes := []string{"sk-", "sk-ant-"}
	hasValidPrefix := false
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(sessionKey, prefix) {
			hasValidPrefix = true
			break
		}
	}
	
	if !hasValidPrefix {
		return false
	}
	
	// 基本格式验证：应该只包含字母、数字、连字符和下划线
	matched, err := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, sessionKey)
	if err != nil || !matched {
		return false
	}
	
	return true
}