package main

import (
	"flag"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/caarlos0/env/v11"
)

var runAddr, logLevel, fileStoragePath string
var storeInterval int
var restore bool

var envSet = map[string]bool{}

func parseFlags() error {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	var flagLogLevel = flag.String("l", "info", "log level")
	var flagStoreInterval = flag.Int("i", 300, "time interval in seconds after which the current server readings are saved to disk")
	var flagFileStoragePath = flag.String("f", "metrics_data", "the path to the file where the current values are saved")
	var flagRestore = flag.Bool("r", false, "whether or not to download previously saved values from the specified file at server startup")

	flag.Parse()

	var cfg config.Config
	err := env.ParseWithOptions(&cfg, env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			envSet[tag] = true
		},
	})
	if err != nil {
		return err
	}

	runAddr = *flagRunAddr
	if (envSet["ADDRESS"]) && (cfg.Address != "") {
		runAddr = cfg.Address
	}

	logLevel = *flagLogLevel
	if (envSet["LOGLEVEL"]) && (cfg.LogLevel != "") {
		logLevel = cfg.LogLevel
	}

	storeInterval = *flagStoreInterval
	if envSet["STORE_INTERVAL"] {
		storeInterval = cfg.StoreInterval
	}

	fileStoragePath = *flagFileStoragePath
	if envSet["FILE_STORAGE_PATH"] && (cfg.FileStoragePath != "") {
		fileStoragePath = cfg.FileStoragePath
	}

	restore = *flagRestore
	if envSet["RESTORE"] {
		restore = cfg.Restore
	}

	return nil
}
