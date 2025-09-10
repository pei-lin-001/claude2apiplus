package config

import (
	"math"
	"sort"
	"sync"
	"time"
)

// SessionHealth Session健康状态
type SessionHealth struct {
	SessionKey      string                 `json:"session_key"`
	OrgID           string                 `json:"org_id"`
	HealthScore     float64                `json:"health_score"`
	Status          SessionStatus          `json:"status"`
	LastUsed        time.Time              `json:"last_used"`
	LastError       time.Time              `json:"last_error,omitempty"`
	CooldownUntil   time.Time              `json:"cooldown_until,omitempty"`
	ErrorCount      int                    `json:"error_count"`
	SuccessCount    int                    `json:"success_count"`
	TotalRequests   int                    `json:"total_requests"`
	AvgResponseTime time.Duration          `json:"avg_response_time"`
	ErrorTypes      map[ErrorType]int      `json:"error_types"`
	RecentErrors    []ErrorRecord          `json:"recent_errors,omitempty"`
	Weights         float64                `json:"weights"`
	CircuitBreaker  *CircuitBreaker        `json:"circuit_breaker,omitempty"`
	mu              sync.RWMutex           `json:"-"`
}

// SessionStatus Session状态
type SessionStatus int

const (
	StatusActive SessionStatus = iota
	StatusCooling
	StatusFailed
	StatusCircuitOpen
)

func (s SessionStatus) String() string {
	switch s {
	case StatusActive:
		return "active"
	case StatusCooling:
		return "cooling"
	case StatusFailed:
		return "failed"
	case StatusCircuitOpen:
		return "circuit_open"
	default:
		return "unknown"
	}
}

// ErrorRecord 错误记录
type ErrorRecord struct {
	Timestamp time.Time `json:"timestamp"`
	ErrorType ErrorType `json:"error_type"`
	Message   string    `json:"message"`
}

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	State         CircuitState `json:"state"`
	FailureCount  int          `json:"failure_count"`
	LastFailure   time.Time    `json:"last_failure"`
	NextAttempt   time.Time    `json:"next_attempt"`
	SuccessCount  int          `json:"success_count"`
	Threshold     int          `json:"threshold"`
	Timeout       time.Duration `json:"timeout"`
}

type CircuitState int

const (
	CircuitClosed CircuitState = iota
	CircuitOpen
	CircuitHalfOpen
)

// SessionManager 智能Session管理器
type SessionManager struct {
	sessions        map[string]*SessionHealth
	config          SessionManagerConfig
	mu              sync.RWMutex
	scheduler       ScheduleStrategy
	stats           *ManagerStats
	lastHealthCheck time.Time
	startTime       time.Time
}

// CallRecord 调用记录
type CallRecord struct {
	SessionKey string    `json:"session_key"`
	Timestamp  time.Time `json:"timestamp"`
	Success    bool      `json:"success"`
	Latency    time.Duration `json:"latency"`
}

// ManagerStats 管理器统计信息
type ManagerStats struct {
	TotalRequests    int64                    `json:"total_requests"`
	SuccessfulReqs   int64                    `json:"successful_requests"`
	FailedRequests   int64                    `json:"failed_requests"`
	AverageLatency   time.Duration            `json:"average_latency"`
	ErrorsByType     map[ErrorType]int64      `json:"errors_by_type"`
	SessionsActive   int                      `json:"sessions_active"`
	SessionsCooling  int                      `json:"sessions_cooling"`
	SessionsFailed   int                      `json:"sessions_failed"`
	LastReset        time.Time                `json:"last_reset"`
	CallRecords      []CallRecord             `json:"call_records"`
	CallCountByHour  map[string]int           `json:"call_count_by_hour"`
	mu               sync.RWMutex             `json:"-"`
}

// ScheduleStrategy 调度策略接口
type ScheduleStrategy interface {
	SelectSession(sessions map[string]*SessionHealth, excludeKeys []string) (*SessionHealth, error)
	GetStrategyName() string
}

// NewSessionManager 创建新的SessionManager
func NewSessionManager(sessions []SessionInfo, config SessionManagerConfig) *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*SessionHealth),
		config:   config,
		stats: &ManagerStats{
			ErrorsByType:    make(map[ErrorType]int64),
			CallRecords:     make([]CallRecord, 0),
			CallCountByHour: make(map[string]int),
			LastReset:       time.Now(),
		},
		lastHealthCheck: time.Now(),
		startTime:       time.Now(),
	}

	// 设置默认配置
	if config.HealthCheckInterval == 0 {
		sm.config.HealthCheckInterval = 30 * time.Second
	}
	if config.MinHealthScore == 0 {
		sm.config.MinHealthScore = 0.5
	}
	if config.MaxRetryAttempts == 0 {
		sm.config.MaxRetryAttempts = 3
	}
	if config.CooldownPeriods == nil {
		sm.config.CooldownPeriods = getDefaultCooldownPeriods()
	}

	// 初始化sessions
	for _, session := range sessions {
		sessionHealth := &SessionHealth{
			SessionKey:      session.SessionKey,
			OrgID:           session.OrgID,
			HealthScore:     1.0,
			Status:          StatusActive,
			LastUsed:        time.Now(),
			ErrorTypes:      make(map[ErrorType]int),
			RecentErrors:    make([]ErrorRecord, 0, 10),
			Weights:         1.0,
		}
		
		// 只有在启用熔断器时才初始化熔断器
		if config.CircuitBreakerEnabled {
			sessionHealth.CircuitBreaker = &CircuitBreaker{
				State:     CircuitClosed,
				Threshold: 5,
				Timeout:   60 * time.Second,
			}
		}
		
		sm.sessions[session.SessionKey] = sessionHealth
	}

	// 选择调度策略
	sm.scheduler = sm.createScheduler()

	return sm
}

// getDefaultCooldownPeriods 获取默认冷却时间
func getDefaultCooldownPeriods() map[string]time.Duration {
	return map[string]time.Duration{
		ErrorRateLimit.String(): 5 * time.Minute,
		ErrorAuth.String():      30 * time.Minute,
		ErrorServer.String():    1 * time.Minute,
		ErrorNetwork.String():   30 * time.Second,
		ErrorTimeout.String():   10 * time.Second,
	}
}

// createScheduler 创建调度策略
func (sm *SessionManager) createScheduler() ScheduleStrategy {
	switch sm.config.ScheduleStrategy {
	case "health_priority":
		return &HealthPriorityStrategy{}
	case "weighted":
		return &WeightedStrategy{}
	case "adaptive":
		return NewAdaptiveStrategy()
	default:
		return &RoundRobinStrategy{}
	}
}

// SelectBestSession 选择最佳session
func (sm *SessionManager) SelectBestSession(excludeKeys []string) (*SessionHealth, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	// 检查是否需要健康检查
	if time.Since(sm.lastHealthCheck) > sm.config.HealthCheckInterval {
		go sm.performHealthCheck()
	}

	// 使用调度策略选择session
	session, err := sm.scheduler.SelectSession(sm.sessions, excludeKeys)
	if err != nil {
		return nil, err
	}

	// 更新使用时间
	if session != nil {
		session.mu.Lock()
		session.LastUsed = time.Now()
		session.mu.Unlock()
	}

	return session, nil
}

// RecordSuccess 记录成功请求
func (sm *SessionManager) RecordSuccess(sessionKey string, responseTime time.Duration) {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionKey]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.SuccessCount++
	session.TotalRequests++
	
	// 更新平均响应时间
	if session.AvgResponseTime == 0 {
		session.AvgResponseTime = responseTime
	} else {
		session.AvgResponseTime = (session.AvgResponseTime + responseTime) / 2
	}

	// 重置熔断器
	if sm.config.CircuitBreakerEnabled && session.CircuitBreaker != nil {
		sm.resetCircuitBreaker(session)
	} else if sm.config.CircuitBreakerEnabled && session.CircuitBreaker == nil {
		// 如果熔断器启用但为空，初始化熔断器
		session.CircuitBreaker = &CircuitBreaker{
			State:     CircuitClosed,
			Threshold: 5,
			Timeout:   60 * time.Second,
		}
	}

	// 重新计算健康度
	sm.updateHealthScore(session)

	// 更新统计
	sm.updateStats(true, responseTime, ErrorOther)
	
	// 记录调用记录
	sm.addCallRecord(sessionKey, true, responseTime)
}

// RecordError 记录错误
func (sm *SessionManager) RecordError(sessionKey string, errorType ErrorType, err error) {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionKey]
	sm.mu.RUnlock()

	if !exists {
		return
	}

	session.mu.Lock()
	defer session.mu.Unlock()

	session.ErrorCount++
	session.TotalRequests++
	session.LastError = time.Now()
	session.ErrorTypes[errorType]++

	// 添加错误记录
	session.RecentErrors = append(session.RecentErrors, ErrorRecord{
		Timestamp: time.Now(),
		ErrorType: errorType,
		Message:   err.Error(),
	})

	// 保持最近错误记录数量
	if len(session.RecentErrors) > 10 {
		session.RecentErrors = session.RecentErrors[1:]
	}

	// 设置冷却时间
	sm.applyCooldown(session, errorType)

	// 更新熔断器
	if sm.config.CircuitBreakerEnabled {
		if session.CircuitBreaker == nil {
			// 如果熔断器启用但为空，初始化熔断器
			session.CircuitBreaker = &CircuitBreaker{
				State:     CircuitClosed,
				Threshold: 5,
				Timeout:   60 * time.Second,
			}
		}
		sm.updateCircuitBreaker(session)
	}

	// 重新计算健康度
	sm.updateHealthScore(session)

	// 更新统计
	sm.updateStats(false, 0, errorType)
	
	// 记录调用记录
	sm.addCallRecord(sessionKey, false, 0)
}

// updateHealthScore 更新健康度评分
func (sm *SessionManager) updateHealthScore(session *SessionHealth) {
	if session.TotalRequests == 0 {
		session.HealthScore = 1.0
		return
	}

	// 基础成功率
	successRate := float64(session.SuccessCount) / float64(session.TotalRequests)
	
	// 错误率影响 (每1%错误率扣0.5分)
	errorRate := float64(session.ErrorCount) / float64(session.TotalRequests)
	errorPenalty := errorRate * 0.5

	// 响应时间影响
	responsePenalty := 0.0
	if session.AvgResponseTime > time.Second {
		seconds := session.AvgResponseTime.Seconds()
		if seconds > 5 {
			responsePenalty = 0.3
		} else if seconds > 2 {
			responsePenalty = 0.2
		} else {
			responsePenalty = 0.1
		}
	}

	// 时间衰减因子 (最近的错误影响更大)
	timeDecay := 1.0
	if !session.LastError.IsZero() {
		hoursSinceError := time.Since(session.LastError).Hours()
		timeDecay = math.Min(1.0, hoursSinceError/24.0) // 24小时内的错误影响更大
	}

	// 状态影响
	statusMultiplier := 1.0
	switch session.Status {
	case StatusCooling:
		statusMultiplier = 0.1
	case StatusFailed, StatusCircuitOpen:
		statusMultiplier = 0.0
	}

	// 计算最终健康度
	baseScore := successRate - errorPenalty - responsePenalty
	session.HealthScore = math.Max(0.0, math.Min(1.0, baseScore*timeDecay*statusMultiplier))
}

// applyCooldown 应用冷却时间
func (sm *SessionManager) applyCooldown(session *SessionHealth, errorType ErrorType) {
	cooldownDuration, exists := sm.config.CooldownPeriods[errorType.String()]
	if !exists {
		cooldownDuration = 1 * time.Minute // 默认冷却时间
	}

	session.CooldownUntil = time.Now().Add(cooldownDuration)
	session.Status = StatusCooling
}

// updateCircuitBreaker 更新熔断器状态
func (sm *SessionManager) updateCircuitBreaker(session *SessionHealth) {
	cb := session.CircuitBreaker
	if cb == nil {
		return
	}

	cb.FailureCount++
	cb.LastFailure = time.Now()

	if cb.State == CircuitClosed && cb.FailureCount >= cb.Threshold {
		cb.State = CircuitOpen
		cb.NextAttempt = time.Now().Add(cb.Timeout)
		session.Status = StatusCircuitOpen
	}
}

// resetCircuitBreaker 重置熔断器
func (sm *SessionManager) resetCircuitBreaker(session *SessionHealth) {
	cb := session.CircuitBreaker
	if cb == nil {
		return
	}

	cb.SuccessCount++
	
	if cb.State == CircuitHalfOpen && cb.SuccessCount >= 3 {
		cb.State = CircuitClosed
		cb.FailureCount = 0
		cb.SuccessCount = 0
		if session.Status == StatusCircuitOpen {
			session.Status = StatusActive
		}
	}
}

// performHealthCheck 执行健康检查
func (sm *SessionManager) performHealthCheck() {
	sm.mu.Lock()
	sm.lastHealthCheck = time.Now()
	sm.mu.Unlock()

	now := time.Now()
	
	for _, session := range sm.sessions {
		session.mu.Lock()
		
		// 检查冷却时间是否结束
		if session.Status == StatusCooling && now.After(session.CooldownUntil) {
			session.Status = StatusActive
		}

		// 检查熔断器是否可以进入半开状态
		if session.Status == StatusCircuitOpen && session.CircuitBreaker != nil {
			if now.After(session.CircuitBreaker.NextAttempt) {
				session.CircuitBreaker.State = CircuitHalfOpen
				session.Status = StatusActive
			}
		}

		session.mu.Unlock()
	}
}

// updateStats 更新统计信息
func (sm *SessionManager) updateStats(success bool, responseTime time.Duration, errorType ErrorType) {
	sm.stats.mu.Lock()
	defer sm.stats.mu.Unlock()

	sm.stats.TotalRequests++
	
	if success {
		sm.stats.SuccessfulReqs++
		// 更新平均延迟
		if sm.stats.AverageLatency == 0 {
			sm.stats.AverageLatency = responseTime
		} else {
			sm.stats.AverageLatency = (sm.stats.AverageLatency + responseTime) / 2
		}
	} else {
		sm.stats.FailedRequests++
		sm.stats.ErrorsByType[errorType]++
	}

	// 更新session状态统计
	active, cooling, failed := 0, 0, 0
	for _, session := range sm.sessions {
		switch session.Status {
		case StatusActive:
			active++
		case StatusCooling:
			cooling++
		case StatusFailed, StatusCircuitOpen:
			failed++
		}
	}
	
	sm.stats.SessionsActive = active
	sm.stats.SessionsCooling = cooling
	sm.stats.SessionsFailed = failed
}

// GetSessionsHealth 获取所有session健康状态
func (sm *SessionManager) GetSessionsHealth() []*SessionHealth {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	sessions := make([]*SessionHealth, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		// 创建副本以避免并发问题
		sessionCopy := *session
		sessions = append(sessions, &sessionCopy)
	}

	// 按健康度排序
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].HealthScore > sessions[j].HealthScore
	})

	return sessions
}

// GetStats 获取统计信息
func (sm *SessionManager) GetStats() *ManagerStats {
	sm.stats.mu.RLock()
	defer sm.stats.mu.RUnlock()

	// 返回副本
	statsCopy := *sm.stats
	statsCopy.ErrorsByType = make(map[ErrorType]int64)
	statsCopy.CallRecords = make([]CallRecord, len(sm.stats.CallRecords))
	statsCopy.CallCountByHour = make(map[string]int)
	
	for k, v := range sm.stats.ErrorsByType {
		statsCopy.ErrorsByType[k] = v
	}
	
	copy(statsCopy.CallRecords, sm.stats.CallRecords)
	for k, v := range sm.stats.CallCountByHour {
		statsCopy.CallCountByHour[k] = v
	}

	return &statsCopy
}

// addCallRecord 添加调用记录
func (sm *SessionManager) addCallRecord(sessionKey string, success bool, latency time.Duration) {
	sm.stats.mu.Lock()
	defer sm.stats.mu.Unlock()
	
	// 创建调用记录
	record := CallRecord{
		SessionKey: sessionKey,
		Timestamp:  time.Now(),
		Success:    success,
		Latency:    latency,
	}
	
	// 添加到调用记录列表
	sm.stats.CallRecords = append(sm.stats.CallRecords, record)
	
	// 保持最近1000条记录
	if len(sm.stats.CallRecords) > 1000 {
		sm.stats.CallRecords = sm.stats.CallRecords[len(sm.stats.CallRecords)-1000:]
	}
	
	// 更新按小时统计
	hourKey := time.Now().Format("2006-01-02 15:00")
	sm.stats.CallCountByHour[hourKey]++
}

// GetMaxRetryAttempts 获取最大重试次数
func (sm *SessionManager) GetMaxRetryAttempts() int {
	return sm.config.MaxRetryAttempts
}

// GetStartTime 获取启动时间
func (sm *SessionManager) GetStartTime() time.Time {
	return sm.startTime
}

// AddSession 动态添加新的session到SessionManager
func (sm *SessionManager) AddSession(sessionInfo SessionInfo) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查session是否已存在
	if _, exists := sm.sessions[sessionInfo.SessionKey]; exists {
		return
	}

	// 创建新的session健康状态
	sessionHealth := &SessionHealth{
		SessionKey:      sessionInfo.SessionKey,
		OrgID:           sessionInfo.OrgID,
		HealthScore:     1.0,
		Status:          StatusActive,
		LastUsed:        time.Now(),
		ErrorTypes:      make(map[ErrorType]int),
		RecentErrors:    make([]ErrorRecord, 0, 10),
		Weights:         1.0,
	}
	
	// 只有在启用熔断器时才初始化熔断器
	if sm.config.CircuitBreakerEnabled {
		sessionHealth.CircuitBreaker = &CircuitBreaker{
			State:     CircuitClosed,
			Threshold: 5,
			Timeout:   60 * time.Second,
		}
	}
	
	// 添加到sessions映射中
	sm.sessions[sessionInfo.SessionKey] = sessionHealth
}

// RemoveSession 从SessionManager中移除session
func (sm *SessionManager) RemoveSession(sessionKey string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 从sessions映射中删除
	delete(sm.sessions, sessionKey)
}

// IsSessionAvailable 检查session是否可用
func (sm *SessionManager) IsSessionAvailable(sessionKey string) bool {
	sm.mu.RLock()
	session, exists := sm.sessions[sessionKey]
	sm.mu.RUnlock()

	if !exists {
		return false
	}

	session.mu.RLock()
	defer session.mu.RUnlock()

	return session.Status == StatusActive && 
		   session.HealthScore >= sm.config.MinHealthScore &&
		   time.Now().After(session.CooldownUntil)
}