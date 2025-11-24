package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap/zapcore"
)

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
	PprofAddr     string `env:"PPROF_ADDR"`
	TrustedSubnet string `env:"TRUSTED_SUBNET"`
}

// DatabaseConfig содержит настройки базы данных
type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN"`
}

// StorageConfig содержит настройки хранилища
type StorageConfig struct {
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

// AgentConfig содержит настройки агента
type AgentConfig struct {
	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
	RateLimit      int `env:"RATE_LIMIT"`
}

// SecurityConfig содержит настройки безопасности
type SecurityConfig struct {
	Key       string `env:"KEY"`
	CryptoKey string `env:"CRYPTO_KEY"`
}

// AuditConfig содержит настройки аудита
type AuditConfig struct {
	File string `env:"AUDIT_FILE"`
	URL  string `env:"AUDIT_URL"`
}

// Config содержит настройки запуска сервера и агента, считываемые из флагов и переменных окружения.
type Config struct {
	LogLevel string `env:"LOGLEVEL"`

	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Agent    AgentConfig
	Security SecurityConfig
	Audit    AuditConfig
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Server.Address)
	enc.AddString("logLevel", c.LogLevel)
	enc.AddInt("reportInterval", c.Agent.ReportInterval)
	enc.AddInt("pollInterval", c.Agent.PollInterval)
	enc.AddInt("storeInterval", c.Server.StoreInterval)
	enc.AddInt("rateLimit", c.Agent.RateLimit)
	enc.AddString("fileStoragePath", c.Storage.FileStoragePath)
	enc.AddBool("restore", c.Server.Restore)
	enc.AddString("dataBaseDSN", c.Database.DSN)
	enc.AddString("key", c.Security.Key)
	enc.AddString("cryptoKey", c.Security.CryptoKey)
	enc.AddString("pprofAddr", c.Server.PprofAddr)
	enc.AddString("trustedSubnet", c.Server.TrustedSubnet)
	enc.AddString("auditFile", c.Audit.File)
	enc.AddString("auditURL", c.Audit.URL)
	return nil
}

// ServerJSONConfig представляет JSON конфигурацию сервера
type ServerJSONConfig struct {
	Address       string `json:"address"`
	Restore       bool   `json:"restore"`
	StoreInterval string `json:"store_interval"`
	StoreFile     string `json:"store_file"`
	DatabaseDSN   string `json:"database_dsn"`
	CryptoKey     string `json:"crypto_key"`
	TrustedSubnet string `json:"trusted_subnet"`
}

// AgentJSONConfig представляет JSON конфигурацию агента
type AgentJSONConfig struct {
	Address        string `json:"address"`
	ReportInterval string `json:"report_interval"`
	PollInterval   string `json:"poll_interval"`
	CryptoKey      string `json:"crypto_key"`
}

// LoadServerConfigFromFile загружает конфигурацию сервера из JSON файла
func LoadServerConfigFromFile(filepath string) (*Config, error) {
	if filepath == "" {
		return &Config{}, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonConfig ServerJSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	config := &Config{}

	if jsonConfig.Address != "" {
		config.Server.Address = jsonConfig.Address
	}

	config.Server.Restore = jsonConfig.Restore

	if jsonConfig.StoreInterval != "" {
		duration, err := time.ParseDuration(jsonConfig.StoreInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid store_interval format: %w", err)
		}
		config.Server.StoreInterval = int(duration.Seconds())
	}

	if jsonConfig.StoreFile != "" {
		config.Storage.FileStoragePath = jsonConfig.StoreFile
	}

	if jsonConfig.DatabaseDSN != "" {
		config.Database.DSN = jsonConfig.DatabaseDSN
	}

	if jsonConfig.CryptoKey != "" {
		config.Security.CryptoKey = jsonConfig.CryptoKey
	}

	return config, nil
}

// LoadAgentConfigFromFile загружает конфигурацию агента из JSON файла
func LoadAgentConfigFromFile(filepath string) (*Config, error) {
	if filepath == "" {
		return &Config{}, nil
	}

	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var jsonConfig AgentJSONConfig
	if err := json.Unmarshal(data, &jsonConfig); err != nil {
		return nil, fmt.Errorf("failed to parse JSON config: %w", err)
	}

	config := &Config{}

	if jsonConfig.Address != "" {
		config.Server.Address = jsonConfig.Address
	}

	if jsonConfig.ReportInterval != "" {
		duration, err := time.ParseDuration(jsonConfig.ReportInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid report_interval format: %w", err)
		}
		config.Agent.ReportInterval = int(duration.Seconds())
	}

	if jsonConfig.PollInterval != "" {
		duration, err := time.ParseDuration(jsonConfig.PollInterval)
		if err != nil {
			return nil, fmt.Errorf("invalid poll_interval format: %w", err)
		}
		config.Agent.PollInterval = int(duration.Seconds())
	}

	if jsonConfig.CryptoKey != "" {
		config.Security.CryptoKey = jsonConfig.CryptoKey
	}

	return config, nil
}

// MergeConfigs объединяет конфигурации с правильным приоритетом:
func MergeConfigs(higher, lower *Config) *Config {
	result := *lower

	if higher.LogLevel != "" {
		result.LogLevel = higher.LogLevel
	}

	// Server config
	if higher.Server.Address != "" {
		result.Server.Address = higher.Server.Address
	}
	if higher.Server.StoreInterval != 0 {
		result.Server.StoreInterval = higher.Server.StoreInterval
	}
	if higher.Server.PprofAddr != "" {
		result.Server.PprofAddr = higher.Server.PprofAddr
	}
	if higher.Server.TrustedSubnet != "" {
		result.Server.TrustedSubnet = higher.Server.TrustedSubnet
	}

	result.Server.Restore = higher.Server.Restore

	// Database config
	if higher.Database.DSN != "" {
		result.Database.DSN = higher.Database.DSN
	}

	// Storage config
	if higher.Storage.FileStoragePath != "" {
		result.Storage.FileStoragePath = higher.Storage.FileStoragePath
	}

	// Agent config
	if higher.Agent.ReportInterval != 0 {
		result.Agent.ReportInterval = higher.Agent.ReportInterval
	}
	if higher.Agent.PollInterval != 0 {
		result.Agent.PollInterval = higher.Agent.PollInterval
	}
	if higher.Agent.RateLimit != 0 {
		result.Agent.RateLimit = higher.Agent.RateLimit
	}

	// Security config
	if higher.Security.Key != "" {
		result.Security.Key = higher.Security.Key
	}
	if higher.Security.CryptoKey != "" {
		result.Security.CryptoKey = higher.Security.CryptoKey
	}

	// Audit config
	if higher.Audit.File != "" {
		result.Audit.File = higher.Audit.File
	}
	if higher.Audit.URL != "" {
		result.Audit.URL = higher.Audit.URL
	}

	return &result
}

// SetDefaultsForServer устанавливает значения по умолчанию для сервера
func SetDefaultsForServer(cfg *Config) {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "localhost:8080"
	}
	if cfg.Server.StoreInterval == 0 {
		cfg.Server.StoreInterval = 300
	}
	if cfg.Storage.FileStoragePath == "" {
		cfg.Storage.FileStoragePath = "metrics_data"
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}

// SetDefaultsForAgent устанавливает значения по умолчанию для агента
func SetDefaultsForAgent(cfg *Config) {
	if cfg.Server.Address == "" {
		cfg.Server.Address = "localhost:8080"
	}
	if cfg.Agent.ReportInterval == 0 {
		cfg.Agent.ReportInterval = 10
	}
	if cfg.Agent.PollInterval == 0 {
		cfg.Agent.PollInterval = 2
	}
	if cfg.LogLevel == "" {
		cfg.LogLevel = "info"
	}
}
