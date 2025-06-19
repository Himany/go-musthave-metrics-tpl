package config

import "go.uber.org/zap/zapcore"

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOGLEVEL"`

	ReportInterval int `env:"REPORT_INTERVAL"`
	PollInterval   int `env:"POLL_INTERVAL"`

	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
}

func (c *Config) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("address", c.Address)
	enc.AddString("logLevel", c.LogLevel)
	enc.AddInt("reportInterval", c.ReportInterval)
	enc.AddInt("pollInterval", c.PollInterval)
	enc.AddInt("storeInterval", c.StoreInterval)
	enc.AddString("fileStoragePath", c.FileStoragePath)
	enc.AddBool("restore", c.Restore)
	enc.AddString("dataBaseDSN", c.DataBaseDSN)
	return nil
}
