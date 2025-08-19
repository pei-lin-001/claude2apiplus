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

type SessionRagen struct {
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
	ChatDelete             bool                 `yaml:"chatDelete"`
	MaxChatHistoryLength   int                  `yaml:"maxChatHistoryLength"`
	RetryCount             int                  `yaml:"retryCount"`
	NoRolePrefix           bool                 `yaml:"noRolePrefix"`
	PromptDisableArtifacts bool                 `yaml:"promptDisableArtifacts"`
	EnableMirrorApi        bool                 `yaml:"enableMirrorApi"`
	MirrorApiPrefix        string               `yaml:"mirrorApiPrefix"`
	RwMutx                 sync.RWMutex         `yaml:"-"` // 不从YAML加载
	sessionManager         *SessionManager      `yaml:"-"` // SessionManager实例
}

// IsSessionManagerEnabled 检查SessionManager是否启用
func (c *Config) IsSessionManagerEnabled() bool {
	return c.SessionManager.Enabled && len(c.Sessions) > 0
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
func (sr *SessionRagen) NextIndex() int {
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

	return &config, nil
}

// 从环境变量加载配置
func loadConfigFromEnv() *Config {
	maxChatHistoryLength, err := strconv.Atoi(os.Getenv("MAX_CHAT_HISTORY_LENGTH"))
	if err != nil {
		maxChatHistoryLength = 10000 // 默认值
	}
	retryCount, sessions := parseSessionEnv(os.Getenv("SESSIONS"))
	config := &Config{
		// 解析 SESSIONS 环境变量
		Sessions: sessions,
		 // SessionManager默认配置（禁用，保持向后兼容）
		SessionManager: SessionManagerConfig{
			Enabled:                false,
			ScheduleStrategy:       "round_robin",
			HealthCheckInterval:    30 * time.Second,
			MinHealthScore:         0.5,
			CircuitBreakerEnabled:  true,
			MaxRetryAttempts:       3,
			CooldownPeriods:        getDefaultCooldownPeriods(),
		},
		// 设置服务地址，默认为 "0.0.0.0:8080"
		Address: os.Getenv("ADDRESS"),

		// 设置 API 认证密钥
		APIKey: os.Getenv("APIKEY"),
		// 设置代理地址
		Proxy: os.Getenv("PROXY"),
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
var Sr *SessionRagen

func init() {
	rand.Seed(time.Now().UnixNano())
	// 加载环境变量
	_ = godotenv.Load()
	Sr = &SessionRagen{
		Index: 0,
		Mutex: sync.Mutex{},
	}
	ConfigInstance = LoadConfig()
	logger.Info("Loaded config:")
	logger.Info(fmt.Sprintf("Max Retry count: %d", ConfigInstance.RetryCount))
	for _, session := range ConfigInstance.Sessions {
		logger.Info(fmt.Sprintf("Session: %s, OrgID: %s", session.SessionKey, session.OrgID))
	}
	logger.Info(fmt.Sprintf("Address: %s", ConfigInstance.Address))
	logger.Info(fmt.Sprintf("APIKey: %s", ConfigInstance.APIKey))
	logger.Info(fmt.Sprintf("Proxy: %s", ConfigInstance.Proxy))
	logger.Info(fmt.Sprintf("ChatDelete: %t", ConfigInstance.ChatDelete))
	logger.Info(fmt.Sprintf("MaxChatHistoryLength: %d", ConfigInstance.MaxChatHistoryLength))
	logger.Info(fmt.Sprintf("NoRolePrefix: %t", ConfigInstance.NoRolePrefix))
	logger.Info(fmt.Sprintf("PromptDisableArtifacts: %t", ConfigInstance.PromptDisableArtifacts))
	logger.Info(fmt.Sprintf("EnableMirrorApi: %t", ConfigInstance.EnableMirrorApi))
	logger.Info(fmt.Sprintf("MirrorApiPrefix: %s", ConfigInstance.MirrorApiPrefix))
}
