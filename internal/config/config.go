package config

import "go.uber.org/zap/zapcore"

// ServerConfig содержит настройки сервера
type ServerConfig struct {
	Address       string `env:"ADDRESS"`
	StoreInterval int    `env:"STORE_INTERVAL"`
	Restore       bool   `env:"RESTORE"`
	PprofAddr     string `env:"PPROF_ADDR"`
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
	enc.AddString("auditFile", c.Audit.File)
	enc.AddString("auditURL", c.Audit.URL)
	return nil
}
