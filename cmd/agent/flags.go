package main

import (
	"flag"
	"fmt"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/caarlos0/env/v11"
)

var envSet = map[string]bool{}

func parseConfig() (*config.Config, error) {
	//Стандратные значения
	const defaultRunAddr = "localhost:8080"
	const defaultReportSeconds = 10
	const defaultPollSeconds = 2
	const defaultLogLevel = "info"

	var flagRunAddr = flag.String("a", defaultRunAddr, "address and port to run server")
	var flagReportSeconds = flag.Int("r", defaultReportSeconds, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", defaultPollSeconds, "poll interval in seconds")
	var flagLogLevel = flag.String("l", defaultLogLevel, "log level")

	flag.Parse()

	var cfg config.Config
	err := env.ParseWithOptions(&cfg, env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			envSet[tag] = true
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing env: %w", err)
	}

	if cfg.Address == "" {
		cfg.Address = *flagRunAddr
	}
	cfg.Address = "http://" + cfg.Address

	if !envSet["REPORT_INTERVAL"] {
		cfg.ReportInterval = *flagReportSeconds
	}

	if envSet["POLL_INTERVAL"] {
		cfg.PollInterval = *flagPollSeconds
	}

	if cfg.LogLevel == "" {
		cfg.LogLevel = *flagLogLevel
	}

	return &cfg, nil
}
