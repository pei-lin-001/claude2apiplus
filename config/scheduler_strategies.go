package config

import (
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// RoundRobinStrategy 轮询调度策略
type RoundRobinStrategy struct {
	currentIndex int
	mu           sync.Mutex
}

func (r *RoundRobinStrategy) SelectSession(sessions map[string]*SessionHealth, excludeKeys []string) (*SessionHealth, error) {
	// 获取可用的sessions
	availableSessions := r.getAvailableSessions(sessions, excludeKeys)
	if len(availableSessions) == 0 {
		return nil, errors.New("no available sessions")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 轮询选择
	session := availableSessions[r.currentIndex%len(availableSessions)]
	r.currentIndex++

	return session, nil
}

func (r *RoundRobinStrategy) GetStrategyName() string {
	return "round_robin"
}

func (r *RoundRobinStrategy) getAvailableSessions(sessions map[string]*SessionHealth, excludeKeys []string) []*SessionHealth {
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}

	var available []*SessionHealth
	for _, session := range sessions {
		if !excludeMap[session.SessionKey] && r.isSessionUsable(session) {
			available = append(available, session)
		}
	}

	// 按SessionKey排序以保证顺序一致性
	sort.Slice(available, func(i, j int) bool {
		return available[i].SessionKey < available[j].SessionKey
	})

	return available
}

func (r *RoundRobinStrategy) isSessionUsable(session *SessionHealth) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()

	return session.Status == StatusActive && 
		   session.HealthScore > 0.1 &&
		   time.Now().After(session.CooldownUntil)
}

// HealthPriorityStrategy 健康度优先调度策略
type HealthPriorityStrategy struct{}

func (h *HealthPriorityStrategy) SelectSession(sessions map[string]*SessionHealth, excludeKeys []string) (*SessionHealth, error) {
	// 获取可用的sessions
	availableSessions := h.getAvailableSessions(sessions, excludeKeys)
	if len(availableSessions) == 0 {
		return nil, errors.New("no available sessions")
	}

	// 按健康度排序，选择最健康的
	sort.Slice(availableSessions, func(i, j int) bool {
		return availableSessions[i].HealthScore > availableSessions[j].HealthScore
	})

	return availableSessions[0], nil
}

func (h *HealthPriorityStrategy) GetStrategyName() string {
	return "health_priority"
}

func (h *HealthPriorityStrategy) getAvailableSessions(sessions map[string]*SessionHealth, excludeKeys []string) []*SessionHealth {
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}

	var available []*SessionHealth
	for _, session := range sessions {
		if !excludeMap[session.SessionKey] && h.isSessionUsable(session) {
			available = append(available, session)
		}
	}

	return available
}

func (h *HealthPriorityStrategy) isSessionUsable(session *SessionHealth) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()

	return session.Status == StatusActive && 
		   session.HealthScore > 0.3 && // 健康度优先策略要求更高的健康度
		   time.Now().After(session.CooldownUntil)
}

// WeightedStrategy 加权调度策略
type WeightedStrategy struct{}

func (w *WeightedStrategy) SelectSession(sessions map[string]*SessionHealth, excludeKeys []string) (*SessionHealth, error) {
	// 获取可用的sessions
	availableSessions := w.getAvailableSessions(sessions, excludeKeys)
	if len(availableSessions) == 0 {
		return nil, errors.New("no available sessions")
	}

	// 计算总权重
	totalWeight := 0.0
	for _, session := range availableSessions {
		session.mu.RLock()
		weight := session.HealthScore * session.Weights
		session.mu.RUnlock()
		totalWeight += weight
	}

	if totalWeight == 0 {
		// 如果总权重为0，退化为轮询
		return availableSessions[rand.Intn(len(availableSessions))], nil
	}

	// 加权随机选择
	randomWeight := rand.Float64() * totalWeight
	currentWeight := 0.0

	for _, session := range availableSessions {
		session.mu.RLock()
		weight := session.HealthScore * session.Weights
		session.mu.RUnlock()
		
		currentWeight += weight
		if currentWeight >= randomWeight {
			return session, nil
		}
	}

	// 备选方案
	return availableSessions[len(availableSessions)-1], nil
}

func (w *WeightedStrategy) GetStrategyName() string {
	return "weighted"
}

func (w *WeightedStrategy) getAvailableSessions(sessions map[string]*SessionHealth, excludeKeys []string) []*SessionHealth {
	excludeMap := make(map[string]bool)
	for _, key := range excludeKeys {
		excludeMap[key] = true
	}

	var available []*SessionHealth
	for _, session := range sessions {
		if !excludeMap[session.SessionKey] && w.isSessionUsable(session) {
			available = append(available, session)
		}
	}

	return available
}

func (w *WeightedStrategy) isSessionUsable(session *SessionHealth) bool {
	session.mu.RLock()
	defer session.mu.RUnlock()

	return session.Status == StatusActive && 
		   session.HealthScore > 0.1 &&
		   time.Now().After(session.CooldownUntil)
}

// AdaptiveStrategy 自适应调度策略 - 根据当前系统状态动态选择策略
type AdaptiveStrategy struct {
	strategies       []ScheduleStrategy
	currentStrategy  int
	lastSwitchTime   time.Time
	performanceStats map[string]*StrategyStats
	mu               sync.RWMutex
}

type StrategyStats struct {
	SuccessCount  int64
	ErrorCount    int64
	AvgLatency    time.Duration
	LastUsed      time.Time
}

func NewAdaptiveStrategy() *AdaptiveStrategy {
	return &AdaptiveStrategy{
		strategies: []ScheduleStrategy{
			&RoundRobinStrategy{},
			&HealthPriorityStrategy{},
			&WeightedStrategy{},
		},
		performanceStats: make(map[string]*StrategyStats),
		lastSwitchTime:   time.Now(),
	}
}

func (a *AdaptiveStrategy) SelectSession(sessions map[string]*SessionHealth, excludeKeys []string) (*SessionHealth, error) {
	a.mu.RLock()
	currentStrategy := a.strategies[a.currentStrategy]
	a.mu.RUnlock()

	// 检查是否需要切换策略
	if time.Since(a.lastSwitchTime) > 5*time.Minute {
		go a.evaluateAndSwitchStrategy(sessions)
	}

	return currentStrategy.SelectSession(sessions, excludeKeys)
}

func (a *AdaptiveStrategy) GetStrategyName() string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return "adaptive_" + a.strategies[a.currentStrategy].GetStrategyName()
}

func (a *AdaptiveStrategy) evaluateAndSwitchStrategy(sessions map[string]*SessionHealth) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 简单的策略评估逻辑
	bestStrategy := 0
	bestPerformance := 0.0

	for i, strategy := range a.strategies {
		stats, exists := a.performanceStats[strategy.GetStrategyName()]
		if !exists {
			continue
		}

		// 计算性能分数
		if stats.SuccessCount+stats.ErrorCount > 0 {
			successRate := float64(stats.SuccessCount) / float64(stats.SuccessCount+stats.ErrorCount)
			latencyPenalty := 1.0
			if stats.AvgLatency > time.Second {
				latencyPenalty = 0.8
			}
			performance := successRate * latencyPenalty
			
			if performance > bestPerformance {
				bestPerformance = performance
				bestStrategy = i
			}
		}
	}

	if bestStrategy != a.currentStrategy {
		a.currentStrategy = bestStrategy
		a.lastSwitchTime = time.Now()
	}
}

func (a *AdaptiveStrategy) RecordResult(success bool, latency time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	strategyName := a.strategies[a.currentStrategy].GetStrategyName()
	stats, exists := a.performanceStats[strategyName]
	if !exists {
		stats = &StrategyStats{}
		a.performanceStats[strategyName] = stats
	}

	if success {
		stats.SuccessCount++
	} else {
		stats.ErrorCount++
	}

	// 更新平均延迟
	if stats.AvgLatency == 0 {
		stats.AvgLatency = latency
	} else {
		stats.AvgLatency = (stats.AvgLatency + latency) / 2
	}
	
	stats.LastUsed = time.Now()
}