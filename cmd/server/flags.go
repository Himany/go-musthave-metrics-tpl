package main

import (
	"flag"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/utils"
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

	utils.SetStringIfUnset(envSet, "ADDRESS", &cfg.Address, *flagRunAddr)
	utils.SetStringIfUnset(envSet, "LOGLEVEL", &cfg.LogLevel, *flagLogLevel)
	utils.SetIntIfUnset(envSet, "STORE_INTERVAL", &cfg.StoreInterval, *flagStoreInterval)
	utils.SetStringIfUnset(envSet, "FILE_STORAGE_PATH", &cfg.FileStoragePath, *flagFileStoragePath)
	utils.SetBoolIfUnset(envSet, "RESTORE", &cfg.Restore, *flagRestore)
	utils.SetStringIfUnset(envSet, "DATABASE_DSN", &cfg.DataBaseDSN, *flagDataBaseDSN)
	utils.SetStringIfUnset(envSet, "KEY", &cfg.Key, *flagKey)

	return &cfg, nil
}
