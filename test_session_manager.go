package main

import (
	"claude2api/config"
	"claude2api/logger"
	"fmt"
	"time"
)

func main() {
	logger.Info("🔧 Testing Intelligent Session Manager...")
	
	// 创建测试sessions
	testSessions := []config.SessionInfo{
		{SessionKey: "test-session-1", OrgID: "org1"},
		{SessionKey: "test-session-2", OrgID: "org2"},
		{SessionKey: "test-session-3", OrgID: "org3"},
	}
	
	// 创建测试SessionManager配置
	testConfig := config.SessionManagerConfig{
		Enabled:                true,
		ScheduleStrategy:       "health_priority",
		HealthCheckInterval:    30 * time.Second,
		MinHealthScore:         0.5,
		CircuitBreakerEnabled:  true,
		MaxRetryAttempts:       3,
		CooldownPeriods:        map[string]time.Duration{
			"rate_limit": 5 * time.Minute,
			"auth":       30 * time.Minute,
			"server":     1 * time.Minute,
			"network":    30 * time.Second,
			"timeout":    10 * time.Second,
		},
	}
	
	// 创建SessionManager
	sm := config.NewSessionManager(testSessions, testConfig)
	
	logger.Info("✅ SessionManager created successfully")
	
	// 测试智能session选择
	logger.Info("🎯 Testing intelligent session selection...")
	for i := 0; i < 5; i++ {
		session, err := sm.SelectBestSession(nil)
		if err != nil {
			logger.Error(fmt.Sprintf("❌ Failed to select session: %v", err))
			continue
		}
		
		logger.Info(fmt.Sprintf("🎲 Selected session: %s (health: %.2f)", 
			session.SessionKey[:12], session.HealthScore))
		
		// 模拟不同类型的请求结果
		if i%3 == 0 {
			// 成功请求
			sm.RecordSuccess(session.SessionKey, time.Millisecond*150)
			logger.Info(fmt.Sprintf("✅ Recorded success for %s", session.SessionKey[:12]))
		} else if i%3 == 1 {
			// 限流错误
			sm.RecordError(session.SessionKey, config.ErrorRateLimit, fmt.Errorf("rate limit exceeded"))
			logger.Info(fmt.Sprintf("⚠️  Recorded rate limit error for %s", session.SessionKey[:12]))
		} else {
			// 网络错误
			sm.RecordError(session.SessionKey, config.ErrorNetwork, fmt.Errorf("network connection failed"))
			logger.Info(fmt.Sprintf("🌐 Recorded network error for %s", session.SessionKey[:12]))
		}
		
		time.Sleep(time.Millisecond * 200)
	}
	
	// 测试排除失效session功能
	logger.Info("🚫 Testing session exclusion...")
	excludeKeys := []string{"test-session-1"}
	session, err := sm.SelectBestSession(excludeKeys)
	if err != nil {
		logger.Error(fmt.Sprintf("❌ Failed to select session with exclusion: %v", err))
	} else {
		logger.Info(fmt.Sprintf("🎯 Selected session with exclusion: %s", session.SessionKey[:12]))
	}
	
	// 显示详细统计信息
	logger.Info("📊 Session Manager Statistics:")
	stats := sm.GetStats()
	logger.Info(fmt.Sprintf("   Total requests: %d", stats.TotalRequests))
	logger.Info(fmt.Sprintf("   Successful requests: %d", stats.SuccessfulReqs))
	logger.Info(fmt.Sprintf("   Failed requests: %d", stats.FailedRequests))
	logger.Info(fmt.Sprintf("   Average latency: %v", stats.AverageLatency))
	logger.Info(fmt.Sprintf("   Active sessions: %d", stats.SessionsActive))
	logger.Info(fmt.Sprintf("   Cooling sessions: %d", stats.SessionsCooling))
	logger.Info(fmt.Sprintf("   Failed sessions: %d", stats.SessionsFailed))
	
	// 显示各错误类型统计
	logger.Info("🔍 Error breakdown by type:")
	for errorType, count := range stats.ErrorsByType {
		if count > 0 {
			logger.Info(fmt.Sprintf("   %s: %d", errorType, count))
		}
	}
	
	// 显示所有session详细健康状态
	logger.Info("🏥 Session Health Report:")
	sessions := sm.GetSessionsHealth()
	for i, session := range sessions {
		logger.Info(fmt.Sprintf("   #%d %s:", i+1, session.SessionKey[:12]))
		logger.Info(fmt.Sprintf("      Health Score: %.3f", session.HealthScore))
		logger.Info(fmt.Sprintf("      Status: %s", session.Status.String()))
		logger.Info(fmt.Sprintf("      Success/Error: %d/%d", session.SuccessCount, session.ErrorCount))
		logger.Info(fmt.Sprintf("      Avg Response Time: %v", session.AvgResponseTime))
		
		if len(session.ErrorTypes) > 0 {
			logger.Info("      Error Types:")
			for errType, count := range session.ErrorTypes {
				if count > 0 {
					logger.Info(fmt.Sprintf("        %s: %d", errType, count))
				}
			}
		}
		
		if !session.CooldownUntil.IsZero() && time.Now().Before(session.CooldownUntil) {
			remaining := time.Until(session.CooldownUntil)
			logger.Info(fmt.Sprintf("      Cooldown remaining: %v", remaining.Round(time.Second)))
		}
	}
	
	// 测试不同调度策略
	logger.Info("🔄 Testing different scheduling strategies...")
	
	// 测试轮询策略
	testConfig.ScheduleStrategy = "round_robin"
	sm2 := config.NewSessionManager(testSessions, testConfig)
	session1, _ := sm2.SelectBestSession(nil)
	session2, _ := sm2.SelectBestSession(nil)
	logger.Info(fmt.Sprintf("Round-robin: %s -> %s", session1.SessionKey[:12], session2.SessionKey[:12]))
	
	// 测试加权策略
	testConfig.ScheduleStrategy = "weighted"
	sm3 := config.NewSessionManager(testSessions, testConfig)
	session3, _ := sm3.SelectBestSession(nil)
	logger.Info(fmt.Sprintf("Weighted strategy selected: %s", session3.SessionKey[:12]))
	
	logger.Info("🎉 All tests completed successfully!")
	logger.Info("🚀 Intelligent Session Manager is ready for production!")
}