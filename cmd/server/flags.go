package main

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address  string `env:"ADDRESS"`
	LogLevel string `env:"LOGLEVEL"`
}

var runAddr, logLevel string

func parseFlags() error {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	var flagLogLevel = flag.String("l", "info", "log level")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return err
	}

	runAddr = *flagRunAddr
	if cfg.Address != "" {
		runAddr = cfg.Address
	}

	logLevel = *flagLogLevel
	if cfg.LogLevel != "" {
		logLevel = cfg.LogLevel
	}

	return nil
}
