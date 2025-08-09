package main

import (
	"flag"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/caarlos0/env/v11"
)

var envSet = map[string]bool{}

// Стандратные значения
const defaultRunAddr = "localhost:8080"
const defaultLogLevel = "info"
const defaultStoreInterval = 300
const defaultFileStoragePath = "metrics_data"
const defaultRestore = false
const defaultDataBaseDSN = ""

func parseFlags() (*config.Config, error) {
	var flagRunAddr = flag.String("a", defaultRunAddr, "address and port to run server")
	var flagLogLevel = flag.String("l", defaultLogLevel, "log level")
	var flagStoreInterval = flag.Int("i", defaultStoreInterval, "time interval in seconds after which the current server readings are saved to disk")
	var flagFileStoragePath = flag.String("f", defaultFileStoragePath, "the path to the file where the current values are saved")
	var flagRestore = flag.Bool("r", defaultRestore, "whether or not to download previously saved values from the specified file at server startup")
	//host=localhost user=postgres password=123321 dbname=metrics sslmode=disable
	var flagDataBaseDSN = flag.String("d", defaultDataBaseDSN, "A string with settings for connecting the postgresql database")
	var flagKey = flag.String("k", "", "key")

	flag.Parse()

	var cfg config.Config
	err := env.ParseWithOptions(&cfg, env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			envSet[tag] = true
		},
	})
	if err != nil {
		return nil, err
	}

	if !((envSet["ADDRESS"]) && (cfg.Address != "")) {
		cfg.Address = *flagRunAddr
	}

	if !((envSet["LOGLEVEL"]) && (cfg.LogLevel != "")) {
		cfg.LogLevel = *flagLogLevel
	}

	if !envSet["STORE_INTERVAL"] {
		cfg.StoreInterval = *flagStoreInterval
	}

	if !(envSet["FILE_STORAGE_PATH"] && (cfg.FileStoragePath != "")) {
		cfg.FileStoragePath = *flagFileStoragePath
	}

	if !envSet["RESTORE"] {
		cfg.Restore = *flagRestore
	}

	if !((envSet["DATABASE_DSN"]) && (cfg.DataBaseDSN != "")) {
		cfg.DataBaseDSN = *flagDataBaseDSN
	}

	if !((envSet["KEY"]) && (cfg.Key != "")) {
		cfg.Key = *flagKey
	}

	return &cfg, nil
}
