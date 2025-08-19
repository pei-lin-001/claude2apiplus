package utils

import (
	"claude2api/config"
	"net/http"
	"strings"
	"time"
)

// RequestResult 请求结果结构
type RequestResult struct {
	Success      bool                   `json:"success"`
	StatusCode   int                    `json:"status_code"`
	Error        error                  `json:"error,omitempty"`
	ResponseTime time.Duration          `json:"response_time"`
	ErrorType    config.ErrorType       `json:"error_type"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

// ClassifyError 根据状态码和错误信息分类错误类型
func ClassifyError(statusCode int, err error) config.ErrorType {
	// 根据HTTP状态码分类
	switch statusCode {
	case http.StatusTooManyRequests: // 429
		return config.ErrorRateLimit
	case http.StatusUnauthorized: // 401
		return config.ErrorAuth
	case http.StatusForbidden: // 403
		return config.ErrorAuth
	case http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return config.ErrorServer
	}
	
	// 根据错误信息分类
	if err != nil {
		errMsg := strings.ToLower(err.Error())
		
		// 网络相关错误
		if strings.Contains(errMsg, "network") ||
			strings.Contains(errMsg, "connection") ||
			strings.Contains(errMsg, "dns") ||
			strings.Contains(errMsg, "socket") {
			return config.ErrorNetwork
		}
		
		// 超时错误
		if strings.Contains(errMsg, "timeout") ||
			strings.Contains(errMsg, "deadline") ||
			strings.Contains(errMsg, "context canceled") {
			return config.ErrorTimeout
		}
		
		// 认证相关错误
		if strings.Contains(errMsg, "unauthorized") ||
			strings.Contains(errMsg, "forbidden") ||
			strings.Contains(errMsg, "invalid") ||
			strings.Contains(errMsg, "expired") {
			return config.ErrorAuth
		}
		
		// 限流相关错误
		if strings.Contains(errMsg, "rate limit") ||
			strings.Contains(errMsg, "too many requests") {
			return config.ErrorRateLimit
		}
	}
	
	return config.ErrorOther
}

// ShouldStopRetry 根据错误类型判断是否应该停止重试
func ShouldStopRetry(errorType config.ErrorType) bool {
	switch errorType {
	case config.ErrorAuth:
		return true // 认证错误通常不会通过重试解决
	default:
		return false // 其他错误可以尝试重试
	}
}

// CalculateBackoffDelay 计算退避延迟时间
func CalculateBackoffDelay(attempt int, errorType config.ErrorType, baseDuration time.Duration) time.Duration {
	// 基础延迟时间
	if baseDuration == 0 {
		baseDuration = time.Second
	}
	
	// 根据错误类型调整基础延迟
	switch errorType {
	case config.ErrorRateLimit:
		baseDuration = 5 * time.Second // 限流错误需要更长的等待时间
	case config.ErrorServer:
		baseDuration = 2 * time.Second // 服务器错误中等等待时间
	case config.ErrorNetwork, config.ErrorTimeout:
		baseDuration = time.Second // 网络和超时错误较短等待时间
	}
	
	// 指数退避算法
	backoffFactor := 1.5
	delay := float64(baseDuration) * float64(attempt) * backoffFactor
	
	// 设置最大延迟时间（30秒）
	maxDelay := 30 * time.Second
	if time.Duration(delay) > maxDelay {
		return maxDelay
	}
	
	return time.Duration(delay)
}

// CreateSuccessResult 创建成功的请求结果
func CreateSuccessResult(statusCode int, responseTime time.Duration) *RequestResult {
	return &RequestResult{
		Success:      true,
		StatusCode:   statusCode,
		ResponseTime: responseTime,
		Details:      make(map[string]interface{}),
	}
}

// CreateErrorResult 创建错误的请求结果
func CreateErrorResult(statusCode int, err error, responseTime time.Duration) *RequestResult {
	errorType := ClassifyError(statusCode, err)
	
	result := &RequestResult{
		Success:      false,
		StatusCode:   statusCode,
		Error:        err,
		ResponseTime: responseTime,
		ErrorType:    errorType,
		Details:      make(map[string]interface{}),
	}
	
	// 添加错误类型特定的详细信息
	switch errorType {
	case config.ErrorRateLimit:
		result.Details["retry_recommended"] = true
		result.Details["min_wait_time"] = "5m"
	case config.ErrorAuth:
		result.Details["retry_recommended"] = false
		result.Details["action_required"] = "check_credentials"
	case config.ErrorServer:
		result.Details["retry_recommended"] = true
		result.Details["min_wait_time"] = "1m"
	case config.ErrorNetwork, config.ErrorTimeout:
		result.Details["retry_recommended"] = true
		result.Details["min_wait_time"] = "30s"
	}
	
	return result
}

// IsRecoverableError 判断错误是否可恢复
func IsRecoverableError(errorType config.ErrorType) bool {
	switch errorType {
	case config.ErrorAuth:
		return false // 认证错误不可恢复
	case config.ErrorRateLimit, config.ErrorServer, config.ErrorNetwork, config.ErrorTimeout:
		return true // 这些错误通常是临时的，可以恢复
	default:
		return true // 默认认为可恢复
	}
}

// GetErrorDescription 获取错误类型的描述信息
func GetErrorDescription(errorType config.ErrorType) string {
	switch errorType {
	case config.ErrorRateLimit:
		return "Rate limit exceeded - too many requests"
	case config.ErrorAuth:
		return "Authentication failed - invalid credentials"
	case config.ErrorServer:
		return "Server error - temporary service unavailable"
	case config.ErrorNetwork:
		return "Network error - connection issues"
	case config.ErrorTimeout:
		return "Timeout error - request took too long"
	default:
		return "Unknown error occurred"
	}
}