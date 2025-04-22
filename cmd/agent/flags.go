package main

import (
	"flag"
	"fmt"
	"strconv"

	"github.com/caarlos0/env/v6"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	Address        string `env:"ADDRESS"`
	ReportInterval string `env:"REPORT_INTERVAL"`
	PollInterval   string `env:"POLL_INTERVAL"`
}

func parseConfig() (*AgentConfig, error) {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	var flagReportSeconds = flag.Int("r", 10, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", 2, "poll interval in seconds")

	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		return nil, fmt.Errorf("error parsing env: %w", err)
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

	return &AgentConfig{
		URL:            "http://" + runAddr + "/update",
		PollInterval:   pollInterval,
		ReportInterval: reportInterval,
		Client:         resty.New(),
	}, nil
}
