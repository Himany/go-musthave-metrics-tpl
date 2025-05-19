package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
	LogLevel       string `env:"LOGLEVEL"`
}

func parseConfig() (string, int, int, string, error) {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	var flagReportSeconds = flag.Int("r", 10, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", 2, "poll interval in seconds")
	var flagLogLevel = flag.String("l", "info", "log level")

	if flagRunAddr == nil || flagReportSeconds == nil || flagPollSeconds == nil || flagLogLevel == nil {
		return "", 0, 0, "", fmt.Errorf("flags init error")
	}

	flag.Parse()

	var cfg Config
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
