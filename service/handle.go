package service

import (
	"claude2api/config"
	"claude2api/core"
	"claude2api/logger"
	"claude2api/model"
	"claude2api/utils"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

// HealthCheckHandler handles the health check endpoint
func HealthCheckHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func MoudlesHandler(c *gin.Context) {
	models := []map[string]interface{}{
		{"id": "claude-3-7-sonnet-20250219"},
		{"id": "claude-sonnet-4-20250514"},
		{"id": "claude-opus-4-20250514"},
	}

	extendedModels := make([]map[string]interface{}, 0, len(models)*2)
	for _, m := range models {
		// 保留原有 id
		extendedModels = append(extendedModels, m)
		// 追加 -think 版本
		if id, ok := m["id"].(string); ok {
			extendedModels = append(extendedModels, map[string]interface{}{
				"id": id + "-think",
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": extendedModels,
	})
}

// ChatCompletionsHandler handles the chat completions endpoint
func ChatCompletionsHandler(c *gin.Context) {
	useMirror, exist := c.Get("UseMirrorApi")
	if exist && useMirror.(bool) {
		MirrorChatHandler(c)
		return
	}

	// Parse and validate request
	req, err := parseAndValidateRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Process messages into prompt and extract images
	processor := utils.NewChatRequestProcessor()
	processor.ProcessMessages(req.Messages)

	// Get model or use default
	model := getModelOrDefault(req.Model)

	// 检查是否启用智能Session管理器
	if config.ConfigInstance.IsSessionManagerEnabled() {
		handleIntelligentChatRequest(c, model, processor, req.Stream)
	} else {
		handleLegacyChatRequest(c, model, processor, req.Stream)
	}
}

// handleIntelligentChatRequest 使用智能Session管理器处理请求
func handleIntelligentChatRequest(c *gin.Context, model string, processor *utils.ChatRequestProcessor, stream bool) {
	sessionManager := config.ConfigInstance.GetSessionManager()
	
	var lastError error
	excludeKeys := make([]string, 0)
	
	// 智能重试循环
	for attempt := 0; attempt < sessionManager.GetMaxRetryAttempts(); attempt++ {
		// 智能选择最佳Session
		sessionHealth, err := sessionManager.SelectBestSession(excludeKeys)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to select session: %v", err))
			break
		}
		
		// 转换为SessionInfo格式以兼容现有代码
		session := config.SessionInfo{
			SessionKey: sessionHealth.SessionKey,
			OrgID:      sessionHealth.OrgID,
		}
		
		logger.Info(fmt.Sprintf("Intelligent session selection: %s (health: %.2f, attempt: %d)", 
			session.SessionKey[:12], sessionHealth.HealthScore, attempt+1))
		
		// 重置processor（如果是重试）
		if attempt > 0 {
			processor.Prompt.Reset()
			processor.Prompt.WriteString(processor.RootPrompt.String())
		}
		
		// 执行请求并收集详细结果
		result := executeRequestWithMetrics(c, session, model, processor, stream)
		
		if result.Success {
			// 记录成功
			sessionManager.RecordSuccess(session.SessionKey, result.ResponseTime)
			logger.Info(fmt.Sprintf("Request successful with session %s in %v", 
				session.SessionKey[:12], result.ResponseTime))
			return
		}
		
		// 分析错误类型并更新Session状态
		errorType := utils.ClassifyError(result.StatusCode, result.Error)
		sessionManager.RecordError(session.SessionKey, errorType, result.Error)
		
		logger.Error(fmt.Sprintf("Request failed with session %s: %s (%s)", 
			session.SessionKey[:12], utils.GetErrorDescription(errorType), result.Error))
		
		// 根据错误类型决定是否继续重试
		if utils.ShouldStopRetry(errorType) {
			logger.Info("Stopping retry due to non-recoverable error")
			lastError = result.Error
			break
		}
		
		// 添加到排除列表，避免立即重试同一个session
		excludeKeys = append(excludeKeys, session.SessionKey)
		lastError = result.Error
		
		// 智能退避策略
		if attempt < sessionManager.GetMaxRetryAttempts()-1 {
			backoffDelay := utils.CalculateBackoffDelay(attempt, errorType, time.Second)
			logger.Info(fmt.Sprintf("Backing off for %v before next attempt", backoffDelay))
			time.Sleep(backoffDelay)
		}
	}
	
	// 所有重试失败后的处理
	logger.Error(fmt.Sprintf("All intelligent retry attempts failed, last error: %v", lastError))
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Failed to process request after intelligent retry attempts"})
}

// handleLegacyChatRequest 使用原始逻辑处理请求（向后兼容）
func handleLegacyChatRequest(c *gin.Context, model string, processor *utils.ChatRequestProcessor, stream bool) {
	index := config.Sr.NextIndex()
	
	// 原始重试逻辑
	for i := 0; i < config.ConfigInstance.RetryCount; i++ {
		index = (index + 1) % len(config.ConfigInstance.Sessions)
		session, err := config.ConfigInstance.GetSessionForModel(index)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get session for model %s: %v", model, err))
			logger.Info("Retrying another session")
			continue
		}

		logger.Info(fmt.Sprintf("Using session for model %s: %s", model, session.SessionKey))
		if i > 0 {
			processor.Prompt.Reset()
			processor.Prompt.WriteString(processor.RootPrompt.String())
		}
		
		// Initialize client and process request
		if handleChatRequest(c, session, model, processor, stream) {
			return // Success, exit the retry loop
		}

		// If we're here, the request failed - retry with another session
		logger.Info("Retrying another session")
	}

	logger.Error("Failed for all retries")
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "Failed to process request after multiple attempts"})
}

// executeRequestWithMetrics 执行请求并收集详细指标
func executeRequestWithMetrics(c *gin.Context, session config.SessionInfo, model string, processor *utils.ChatRequestProcessor, stream bool) *utils.RequestResult {
	startTime := time.Now()
	
	// Initialize the Claude client
	claudeClient := core.NewClient(session.SessionKey, config.ConfigInstance.Proxy, model)

	// Get org ID if not already set
	if session.OrgID == "" {
		orgId, err := claudeClient.GetOrgID()
		if err != nil {
			return utils.CreateErrorResult(500, err, time.Since(startTime))
		}
		session.OrgID = orgId
		config.ConfigInstance.SetSessionOrgID(session.SessionKey, session.OrgID)
	}

	claudeClient.SetOrgID(session.OrgID)

	// Upload images if any
	if len(processor.ImgDataList) > 0 {
		err := claudeClient.UploadFile(processor.ImgDataList)
		if err != nil {
			return utils.CreateErrorResult(500, err, time.Since(startTime))
		}
	}

	// Handle large context if needed
	if processor.Prompt.Len() > config.ConfigInstance.MaxChatHistoryLength {
		claudeClient.SetBigContext(processor.Prompt.String())
		processor.ResetForBigContext()
		logger.Info(fmt.Sprintf("Prompt length exceeds max limit (%d), using file context", config.ConfigInstance.MaxChatHistoryLength))
	}

	// Create conversation
	conversationID, err := claudeClient.CreateConversation()
	if err != nil {
		return utils.CreateErrorResult(500, err, time.Since(startTime))
	}

	// Send message
	statusCode, err := claudeClient.SendMessage(conversationID, processor.Prompt.String(), stream, c)
	responseTime := time.Since(startTime)
	
	if err != nil {
		// Cleanup conversation asynchronously
		go cleanupConversation(claudeClient, conversationID, 3)
		return utils.CreateErrorResult(statusCode, err, responseTime)
	}

	// Clean up conversation if enabled
	if config.ConfigInstance.ChatDelete {
		go cleanupConversation(claudeClient, conversationID, 3)
	}

	return utils.CreateSuccessResult(statusCode, responseTime)
}

func MirrorChatHandler(c *gin.Context) {
	if !config.ConfigInstance.EnableMirrorApi {
		c.JSON(http.StatusForbidden, ErrorResponse{
			Error: "Mirror API is not enabled",
		})
		return
	}

	// Parse and validate request
	req, err := parseAndValidateRequest(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	// Process messages into prompt and extract images
	processor := utils.NewChatRequestProcessor()
	processor.ProcessMessages(req.Messages)

	// Get model or use default
	model := getModelOrDefault(req.Model)

	// Extract session info from auth header
	session, err := extractSessionFromAuthHeader(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid authorization: %v", err),
		})
		return
	}

	// Process the request with the provided session
	if !handleChatRequest(c, session, model, processor, req.Stream) {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error: "Failed to process request",
		})
		return
	}
}

// Helper functions

func parseAndValidateRequest(c *gin.Context) (*model.ChatCompletionRequest, error) {
	var req model.ChatCompletionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: fmt.Sprintf("Invalid request: %v", err),
		})
		return nil, err
	}

	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "No messages provided",
		})
		return nil, fmt.Errorf("no messages provided")
	}

	return &req, nil
}

func getModelOrDefault(model string) string {
	if model == "" {
		return "claude-3-7-sonnet-20250219"
	}
	return model
}

func extractSessionFromAuthHeader(c *gin.Context) (config.SessionInfo, error) {
	authInfo := c.Request.Header.Get("Authorization")
	authInfo = strings.TrimPrefix(authInfo, "Bearer ")

	if authInfo == "" {
		return config.SessionInfo{SessionKey: "", OrgID: ""}, fmt.Errorf("missing authorization header")
	}

	if strings.Contains(authInfo, ":") {
		parts := strings.Split(authInfo, ":")
		return config.SessionInfo{SessionKey: parts[0], OrgID: parts[1]}, nil
	}

	return config.SessionInfo{SessionKey: authInfo, OrgID: ""}, nil
}

func handleChatRequest(c *gin.Context, session config.SessionInfo, model string, processor *utils.ChatRequestProcessor, stream bool) bool {
	// Initialize the Claude client
	claudeClient := core.NewClient(session.SessionKey, config.ConfigInstance.Proxy, model)

	// Get org ID if not already set
	if session.OrgID == "" {
		orgId, err := claudeClient.GetOrgID()
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to get org ID: %v", err))
			return false
		}
		session.OrgID = orgId
		config.ConfigInstance.SetSessionOrgID(session.SessionKey, session.OrgID)
	}

	claudeClient.SetOrgID(session.OrgID)

	// Upload images if any
	if len(processor.ImgDataList) > 0 {
		err := claudeClient.UploadFile(processor.ImgDataList)
		if err != nil {
			logger.Error(fmt.Sprintf("Failed to upload file: %v", err))
			return false
		}
	}

	// Handle large context if needed
	if processor.Prompt.Len() > config.ConfigInstance.MaxChatHistoryLength {
		claudeClient.SetBigContext(processor.Prompt.String())
		processor.ResetForBigContext()
		logger.Info(fmt.Sprintf("Prompt length exceeds max limit (%d), using file context", config.ConfigInstance.MaxChatHistoryLength))
	}

	// Create conversation
	conversationID, err := claudeClient.CreateConversation()
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to create conversation: %v", err))
		return false
	}

	// Send message
	if _, err := claudeClient.SendMessage(conversationID, processor.Prompt.String(), stream, c); err != nil {
		logger.Error(fmt.Sprintf("Failed to send message: %v", err))
		go cleanupConversation(claudeClient, conversationID, 3)
		return false
	}

	// Clean up conversation if enabled
	if config.ConfigInstance.ChatDelete {
		go cleanupConversation(claudeClient, conversationID, 3)
	}

	return true
}

func cleanupConversation(client *core.Client, conversationID string, retry int) {
	for i := 0; i < retry; i++ {
		if err := client.DeleteConversation(conversationID); err != nil {
			logger.Error(fmt.Sprintf("Failed to delete conversation: %v", err))
			time.Sleep(2 * time.Second)
			continue
		}
		logger.Info(fmt.Sprintf("Successfully deleted conversation: %s", conversationID))
		return // 成功后直接返回，不执行后面的错误日志
	}
	// 只有当所有重试都失败后，才会执行到这里
	logger.Error(fmt.Sprintf("Cleanup %s conversation %s failed after %d retries", client.SessionKey, conversationID, retry))
}
