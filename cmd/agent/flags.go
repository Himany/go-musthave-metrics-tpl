package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/caarlos0/env/v11"
)

func parseConfig() (string, int, int, string, error) {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	var flagReportSeconds = flag.Int("r", 10, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", 2, "poll interval in seconds")
	var flagLogLevel = flag.String("l", "info", "log level")

	flag.Parse()

	var cfg config.Config
	err := env.Parse(&cfg)
	if err != nil {
		return "", 0, 0, "", fmt.Errorf("error parsing env: %w", err)
	}

	runAddr := *flagRunAddr
	if cfg.Address != "" {
		runAddr = cfg.Address
	}

	reportInterval := *flagReportSeconds
	if cfg.ReportInterval != "" {
		if v, err := strconv.Atoi(cfg.ReportInterval); err == nil {
			reportInterval = v
		}
	}

	pollInterval := *flagPollSeconds
	if cfg.PollInterval != "" {
		if v, err := strconv.Atoi(cfg.PollInterval); err == nil {
			pollInterval = v
		}
	}

	logLvl := *flagLogLevel
	if cfg.LogLevel != "" {
		logLvl = cfg.LogLevel
	}

	return "http://" + runAddr, reportInterval, pollInterval, logLvl, nil
}
