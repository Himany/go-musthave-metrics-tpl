package config

import "go.uber.org/zap/zapcore"

// Config содержит настройки запуска сервера и агента, считываемые из флагов и переменных окружения.
type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOGLEVEL"`

	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`
	RateLimit      int `env:"RATE_LIMIT"`

	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataBaseDSN     string `env:"DATABASE_DSN"`

	Key string `env:"KEY"`

	AuditFile string `env:"AUDIT_FILE"`
	AuditURL  string `env:"AUDIT_URL"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddString("logLevel", c.LogLevel)
	enc.AddInt("reportInterval", c.ReportInterval)
	enc.AddInt("pollInterval", c.PollInterval)
	enc.AddInt("storeInterval", c.StoreInterval)
	enc.AddInt("rateLimit", c.RateLimit)
	enc.AddString("fileStoragePath", c.FileStoragePath)
	enc.AddBool("restore", c.Restore)
	enc.AddString("dataBaseDSN", c.DataBaseDSN)
	enc.AddString("key", c.Key)
	return nil
}
