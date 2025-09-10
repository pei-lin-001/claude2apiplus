package config

import (
	"claude2api/logger"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// ErrorType 错误类型枚举
type ErrorType int

const (
	ErrorOther ErrorType = iota
	ErrorRateLimit
	ErrorAuth
	ErrorServer
	ErrorNetwork
	ErrorTimeout
)

// String 实现String接口
func (e ErrorType) String() string {
	switch e {
	case ErrorRateLimit:
		return "rate_limit"
	case ErrorAuth:
		return "auth"
	case ErrorServer:
		return "server"
	case ErrorNetwork:
		return "network"
	case ErrorTimeout:
		return "timeout"
	default:
		return "other"
	}
}

type SessionInfo struct {
	SessionKey string `yaml:"sessionKey"`
	OrgID      string `yaml:"orgID"`
}

type SessionRange struct {
	Index int
	Mutex sync.Mutex
}

// SessionManagerConfig SessionManager配置
type SessionManagerConfig struct {
	Enabled                bool                   `yaml:"enabled"`
	ScheduleStrategy       string                `yaml:"scheduleStrategy"` // "round_robin", "health_priority", "weighted"
	HealthCheckInterval    time.Duration         `yaml:"healthCheckInterval"`
	MinHealthScore         float64               `yaml:"minHealthScore"`
	CircuitBreakerEnabled  bool                  `yaml:"circuitBreakerEnabled"`
	MaxRetryAttempts       int                   `yaml:"maxRetryAttempts"`
	CooldownPeriods        map[string]time.Duration `yaml:"cooldownPeriods"`
}

type Config struct {
    Sessions               []SessionInfo        `yaml:"sessions"`
    SessionManager         SessionManagerConfig `yaml:"sessionManager"`
    Address                string               `yaml:"address"`
    APIKey                 string               `yaml:"apiKey"`
    Proxy                  string               `yaml:"proxy"`
    CORSAllowedOrigins     []string             `yaml:"corsAllowedOrigins"`
    ChatDelete             bool                 `yaml:"chatDelete"`
    MaxChatHistoryLength   int                  `yaml:"maxChatHistoryLength"`
    RetryCount             int                  `yaml:"retryCount"`
    NoRolePrefix           bool                 `yaml:"noRolePrefix"`
    PromptDisableArtifacts bool                 `yaml:"promptDisableArtifacts"`
	EnableMirrorApi        bool                 `yaml:"enableMirrorApi"`
	MirrorApiPrefix        string               `yaml:"mirrorApiPrefix"`
	AdminUser              string               `yaml:"adminUser"`
	AdminPassword          string               `yaml:"adminPassword"`
	AdminSecret            string               `yaml:"adminSecret"`
	RwMutx                 sync.RWMutex         `yaml:"-"` // 不从YAML加载
	sessionManager         *SessionManager      `yaml:"-"` // SessionManager实例
}

// IsSessionManagerEnabled 检查SessionManager是否启用
func (c *Config) IsSessionManagerEnabled() bool {
	return c.SessionManager.Enabled && len(c.Sessions) > 0
}

// GetAdminUser 获取管理员用户名
func (c *Config) GetAdminUser() string {
	if c.AdminUser == "" {
		return "admin"
	}
	return c.AdminUser
}

// GetAdminPassword 获取管理员密码
func (c *Config) GetAdminPassword() string {
	if c.AdminPassword == "" {
		return "admin123"
	}
	return c.AdminPassword
}

// GetAdminSecret 获取管理员JWT密钥
func (c *Config) GetAdminSecret() string {
	if c.AdminSecret == "" {
		return "claude2api-admin-secret-key"
	}
	return c.AdminSecret
}

// ValidateConfig 验证配置
func (c *Config) ValidateConfig() error {
	if len(c.Sessions) == 0 {
		return fmt.Errorf("no sessions configured")
	}
	
	for i, session := range c.Sessions {
		if session.SessionKey == "" {
			return fmt.Errorf("session %d has empty session key", i+1)
		}
	}
	
	if c.APIKey == "" {
		return fmt.Errorf("API key is required")
	}
	
	if c.SessionManager.Enabled {
		if c.SessionManager.ScheduleStrategy != "round_robin" && 
		   c.SessionManager.ScheduleStrategy != "health_priority" && 
		   c.SessionManager.ScheduleStrategy != "weighted" &&
		   c.SessionManager.ScheduleStrategy != "adaptive" {
			return fmt.Errorf("invalid session manager strategy: %s", c.SessionManager.ScheduleStrategy)
		}
		
		if c.SessionManager.MinHealthScore < 0 || c.SessionManager.MinHealthScore > 1 {
			return fmt.Errorf("min health score must be between 0 and 1")
		}
		
		if c.SessionManager.MaxRetryAttempts < 0 {
			return fmt.Errorf("max retry attempts must be non-negative")
		}
	}
	
	return nil
}

// GetSessionManager 获取SessionManager实例
func (c *Config) GetSessionManager() *SessionManager {
	if c.sessionManager == nil && c.IsSessionManagerEnabled() {
		c.sessionManager = NewSessionManager(c.Sessions, c.SessionManager)
	}
	return c.sessionManager
}

// 解析 SESSION 格式的环境变量
func parseSessionEnv(envValue string) (int, []SessionInfo) {
	if envValue == "" {
		return 0, []SessionInfo{}
	}
	var sessions []SessionInfo
	sessionPairs := strings.Split(envValue, ",")
	retryCount := len(sessionPairs) // 重试次数等于 session 数量
	for _, pair := range sessionPairs {
		if pair == "" {
			retryCount--
			continue
		}
		parts := strings.Split(pair, ":")
		session := SessionInfo{
			SessionKey: parts[0],
		}

		if len(parts) > 1 {
			session.OrgID = parts[1]
		} else if len(parts) == 1 {
			session.OrgID = ""
		}

		sessions = append(sessions, session)
	}
	if retryCount > 5 {
		retryCount = 5 // 限制最大重试次数为 5 次
	}
	return retryCount, sessions
}

// 根据模型选择合适的 session
func (c *Config) GetSessionForModel(idx int) (SessionInfo, error) {
	if len(c.Sessions) == 0 || idx < 0 || idx >= len(c.Sessions) {
		return SessionInfo{}, fmt.Errorf("invalid session index: %d", idx)
	}
	c.RwMutx.RLock()
	defer c.RwMutx.RUnlock()
	return c.Sessions[idx], nil
}

func (c *Config) SetSessionOrgID(sessionKey, orgID string) {
	c.RwMutx.Lock()
	defer c.RwMutx.Unlock()
	for i, session := range c.Sessions {
		if session.SessionKey == sessionKey {
			logger.Info(fmt.Sprintf("Setting OrgID for session %s to %s", sessionKey, orgID))
			c.Sessions[i].OrgID = orgID
			return
		}
	}
}
func (sr *SessionRange) NextIndex() int {
	sr.Mutex.Lock()
	defer sr.Mutex.Unlock()

	index := sr.Index
	sr.Index = (index + 1) % len(ConfigInstance.Sessions)
	return index
}

// 检查配置文件是否存在
func configFileExists() (bool, string) {
	execDir := filepath.Dir(os.Args[0])
	workDir, _ := os.Getwd()
	if execDir == "" && workDir == "" {
		logger.Error("Failed to get executable directory")
		return false, ""
	}

	var err error
	exeConfigPath := filepath.Join(execDir, "config.yaml")
	_, err = os.Stat(exeConfigPath)
	if !os.IsNotExist(err) {
		return true, exeConfigPath
	}

	workConfigPath := filepath.Join(workDir, "config.yaml")
	_, err = os.Stat(workConfigPath)
	if !os.IsNotExist(err) {
		return true, workConfigPath
	}

	return false, ""
}

// 从YAML文件加载配置
func loadConfigFromYAML(configPath string) (*Config, error) {
    data, err := os.ReadFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read config file: %v", err)
    }

    var config Config
    err = yaml.Unmarshal(data, &config)
    if err != nil {
        return nil, fmt.Errorf("failed to parse config file: %v", err)
    }

    // 设置读写锁（不从YAML加载）
    config.RwMutx = sync.RWMutex{}

    // 如果地址为空，使用默认值
    if config.Address == "" {
        config.Address = "0.0.0.0:8080"
    }

    // CORS 允许来源默认值
    if len(config.CORSAllowedOrigins) == 0 {
        config.CORSAllowedOrigins = []string{"*"}
    }

    return &config, nil
}

// 从环境变量加载配置
func loadConfigFromEnv() *Config {
	maxChatHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_LENGTH"))
	if err != nil {
		maxChatHistoryLength = 10000 // 默认值
	}
	retryCount, sessions := parseSessionEnv(os.Getenv("SESSIONS"))
	
	// 解析SessionManager环境变量
	sessionManagerEnabled := os.Getenv("SESSION_MANAGER_ENABLED") == "true"
	sessionManagerStrategy := os.Getenv("SESSION_MANAGER_STRATEGY")
	if sessionManagerStrategy == "" {
		sessionManagerStrategy = "round_robin"
	}
	
	healthCheckInterval, err := time.ParseDuration(os.Getenv("HEALTH_CHECK_INTERVAL"))
	if err != nil {
		healthCheckInterval = 30 * time.Second
	}
	
	minHealthScore, err := strconv.ParseFloat(os.Getenv("MIN_HEALTH_SCORE"), 64)
	if err != nil {
		minHealthScore = 0.5
	}
	
	circuitBreakerEnabled := os.Getenv("CIRCUIT_BREAKER_ENABLED") != "false"
	maxRetryAttempts, err := strconv.Atoi(os.Getenv("MAX_RETRY_ATTEMPTS"))
	if err != nil {
		maxRetryAttempts = 3
	}
	
    config := &Config{
        // 解析 SESSIONS 环境变量
        Sessions: sessions,
        // SessionManager配置
        SessionManager: SessionManagerConfig{
			Enabled:                sessionManagerEnabled,
			ScheduleStrategy:       sessionManagerStrategy,
			HealthCheckInterval:    healthCheckInterval,
			MinHealthScore:         minHealthScore,
			CircuitBreakerEnabled:  circuitBreakerEnabled,
			MaxRetryAttempts:       maxRetryAttempts,
			CooldownPeriods:        getDefaultCooldownPeriods(),
		},
        // 设置服务地址，默认为 "0.0.0.0:8080"
        Address: os.Getenv("ADDRESS"),

        // 设置 API 认证密钥
        APIKey: os.Getenv("APIKEY"),
        // 设置代理地址
        Proxy: os.Getenv("PROXY"),
        // CORS 允许来源，逗号分隔，默认 *
        CORSAllowedOrigins: func() []string {
            v := os.Getenv("CORS_ORIGINS")
            if strings.TrimSpace(v) == "" {
                return []string{"*"}
            }
            parts := strings.Split(v, ",")
            outs := make([]string, 0, len(parts))
            for _, p := range parts {
                s := strings.TrimSpace(p)
                if s != "" {
                    outs = append(outs, s)
                }
            }
            if len(outs) == 0 {
                return []string{"*"}
            }
            return outs
        }(),
        // 自动删除聊天
        ChatDelete: os.Getenv("CHAT_DELETE") != "false",
        // 设置最大聊天历史长度
        MaxChatHistoryLength: maxChatHistoryLength,
        // 设置重试次数
		RetryCount: retryCount,
		// 设置是否使用角色前缀
		NoRolePrefix: os.Getenv("NO_ROLE_PREFIX") == "true",
		// 设置是否使用提示词禁用artifacts
		PromptDisableArtifacts: os.Getenv("PROMPT_DISABLE_ARTIFACTS") == "true",
		// 设置是否启用镜像API
		EnableMirrorApi: os.Getenv("ENABLE_MIRROR_API") == "true",
		// 设置镜像API前缀
		MirrorApiPrefix: os.Getenv("MIRROR_API_PREFIX"),
		// 设置管理员凭据
		AdminUser: os.Getenv("ADMIN_USER"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
		AdminSecret: os.Getenv("ADMIN_SECRET"),
		// 设置读写锁
		RwMutx: sync.RWMutex{},
	}

	// 如果地址为空，使用默认值
	if config.Address == "" {
		config.Address = "0.0.0.0:8080"
	}
	return config
}

// 加载配置
func LoadConfig() *Config {
	// 检查配置文件是否存在
	exists, configPath := configFileExists()
	if exists {
		logger.Info(fmt.Sprintf("Found config file at %s", configPath))
		config, err := loadConfigFromYAML(configPath)
		if err == nil {
			logger.Info("Successfully loaded configuration from YAML file")
			return config
		}
		logger.Error(fmt.Sprintf("Failed to load config from YAML: %v, falling back to environment variables", err))
	}

	// 如果配置文件不存在或加载失败，从环境变量加载
	logger.Info("Loading configuration from environment variables")
	return loadConfigFromEnv()
}

var ConfigInstance *Config
var Sr *SessionRange

func init() {
	rand.Seed(time.Now().UnixNano())
	// 加载环境变量
	_ = godotenv.Load()
	Sr = &SessionRange{
		Index: 0,
		Mutex: sync.Mutex{},
	}
	ConfigInstance = LoadConfig()
	
	// 验证配置
	if err := ConfigInstance.ValidateConfig(); err != nil {
		logger.Error(fmt.Sprintf("Configuration validation failed: %v", err))
		os.Exit(1)
	}
	
    logger.Info("Loaded config:")
    logger.Info(fmt.Sprintf("Max Retry count: %d", ConfigInstance.RetryCount))
    for _, session := range ConfigInstance.Sessions {
        masked := logger.MaskSecret(session.SessionKey)
        logger.Info(fmt.Sprintf("Session: %s, OrgID: %s", masked, session.OrgID))
    }
    logger.Info(fmt.Sprintf("Address: %s", ConfigInstance.Address))
    if ConfigInstance.APIKey != "" {
        logger.Info(fmt.Sprintf("APIKey: %s", logger.MaskSecret(ConfigInstance.APIKey)))
    }
    if ConfigInstance.Proxy != "" {
        logger.Info(fmt.Sprintf("Proxy: %s", ConfigInstance.Proxy))
    }
    logger.Info(fmt.Sprintf("ChatDelete: %t", ConfigInstance.ChatDelete))
    logger.Info(fmt.Sprintf("MaxChatHistoryLength: %d", ConfigInstance.MaxChatHistoryLength))
    logger.Info(fmt.Sprintf("NoRolePrefix: %t", ConfigInstance.NoRolePrefix))
    logger.Info(fmt.Sprintf("PromptDisableArtifacts: %t", ConfigInstance.PromptDisableArtifacts))
    logger.Info(fmt.Sprintf("EnableMirrorApi: %t", ConfigInstance.EnableMirrorApi))
    logger.Info(fmt.Sprintf("MirrorApiPrefix: %s", ConfigInstance.MirrorApiPrefix))
    logger.Info(fmt.Sprintf("CORS Allowed Origins: %v", ConfigInstance.CORSAllowedOrigins))
    logger.Info(fmt.Sprintf("SessionManager Enabled: %t", ConfigInstance.SessionManager.Enabled))
    logger.Info(fmt.Sprintf("SessionManager Strategy: %s", ConfigInstance.SessionManager.ScheduleStrategy))
    logger.Info(fmt.Sprintf("SessionManager MinHealthScore: %f", ConfigInstance.SessionManager.MinHealthScore))
    logger.Info(fmt.Sprintf("SessionManager MaxRetryAttempts: %d", ConfigInstance.SessionManager.MaxRetryAttempts))

    // 提示默认管理员凭据风险
    if ConfigInstance.GetAdminUser() == "admin" && ConfigInstance.GetAdminPassword() == "admin123" {
        logger.Warn("Admin credentials are default values; please change in production")
    }
}
