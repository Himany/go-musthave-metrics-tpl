package config

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOGLEVEL"`

	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`

	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DataBaseDSN     string `env:"DATABASE_DSN"`
}
