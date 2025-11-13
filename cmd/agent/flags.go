package main

import (
	"flag"
	"fmt"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/utils"
	"github.com/caarlos0/env/v11"
)

// envTracker инкапсулирует информацию о выставленных переменных окружения.
type envTracker map[string]bool

// Стандратные значения
const defaultRunAddr = "localhost:8080"
const defaultReportSeconds = 10
const defaultPollSeconds = 2
const defaultLogLevel = "info"
const defaultRateLimit = 0

func parseConfig() (*config.Config, error) {
	envSet := make(envTracker)
	var flagRunAddr = flag.String("a", defaultRunAddr, "address and port to run server")
	var flagReportSeconds = flag.Int("r", defaultReportSeconds, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", defaultPollSeconds, "poll interval in seconds")
	//var flagLogLevel = flag.String("l", defaultLogLevel, "log level")
	var flagKey = flag.String("k", "", "Key")
	var flagRateLimit = flag.Int("l", defaultRateLimit, "maximum number of simultaneous requests to the server")
	var flagCryptoKey = flag.String("crypto-key", "", "path to public key file for asymmetric encryption")

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

	utils.SetStringIfUnset(envSet, "ADDRESS", &cfg.Server.Address, *flagRunAddr)
	cfg.Server.Address = "http://" + cfg.Server.Address
	utils.SetIntIfUnset(envSet, "REPORT_INTERVAL", &cfg.Agent.ReportInterval, *flagReportSeconds)
	utils.SetIntIfUnset(envSet, "POLL_INTERVAL", &cfg.Agent.PollInterval, *flagPollSeconds)
	utils.SetStringIfUnset(envSet, "LOG_LEVEL", &cfg.LogLevel, defaultLogLevel)
	utils.SetStringIfUnset(envSet, "KEY", &cfg.Security.Key, *flagKey)
	utils.SetIntIfUnset(envSet, "RATE_LIMIT", &cfg.Agent.RateLimit, *flagRateLimit)
	utils.SetStringIfUnset(envSet, "CRYPTO_KEY", &cfg.Security.CryptoKey, *flagCryptoKey)

	return &cfg, nil
}
