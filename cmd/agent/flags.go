package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/utils"
	"github.com/caarlos0/env/v11"
)

// envTracker инкапсулирует информацию о выставленных переменных окружения.
type envTracker map[string]bool

// Стандратные значения
const defaultRunAddr = "localhost:8080"
const defaultGRPCAddr = "localhost:9090"
const defaultReportSeconds = 10
const defaultPollSeconds = 2
const defaultLogLevel = "info"
const defaultRateLimit = 0

func parseConfig() (*config.Config, error) {
	envSet := make(envTracker)
	var flagRunAddr = flag.String("a", defaultRunAddr, "address and port to run server")
	var flagGRPCAddr = flag.String("grpc", defaultGRPCAddr, "address and port of gRPC server")
	var flagReportSeconds = flag.Int("r", defaultReportSeconds, "report interval in seconds")
	var flagPollSeconds = flag.Int("p", defaultPollSeconds, "poll interval in seconds")
	//var flagLogLevel = flag.String("l", defaultLogLevel, "log level")
	var flagKey = flag.String("k", "", "Key")
	var flagRateLimit = flag.Int("l", defaultRateLimit, "maximum number of simultaneous requests to the server")
	var flagCryptoKey = flag.String("crypto-key", "", "path to public key file for asymmetric encryption")
	var flagConfigFile = flag.String("c", "", "path to JSON configuration file")
	var flagConfigFileLong = flag.String("config", "", "path to JSON configuration file")

	flag.Parse()

	configPath := ""
	if *flagConfigFile != "" {
		configPath = *flagConfigFile
	} else if *flagConfigFileLong != "" {
		configPath = *flagConfigFileLong
	} else if envConfigPath := os.Getenv("CONFIG"); envConfigPath != "" {
		configPath = envConfigPath
	}

	configFromFile, err := config.LoadAgentConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	config.SetDefaultsForAgent(configFromFile)

	var envConfig config.Config
	err = env.ParseWithOptions(&envConfig, env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			envSet[tag] = true
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing env: %w", err)
	}

	flagConfig := &config.Config{}

	utils.SetStringIfUnset(envSet, "ADDRESS", &flagConfig.Server.Address, *flagRunAddr)
	utils.SetStringIfUnset(envSet, "GRPC_ADDRESS", &flagConfig.Server.GRPCAddress, *flagGRPCAddr)
	utils.SetIntIfUnset(envSet, "REPORT_INTERVAL", &flagConfig.Agent.ReportInterval, *flagReportSeconds)
	utils.SetIntIfUnset(envSet, "POLL_INTERVAL", &flagConfig.Agent.PollInterval, *flagPollSeconds)
	utils.SetStringIfUnset(envSet, "LOG_LEVEL", &flagConfig.LogLevel, defaultLogLevel)
	utils.SetStringIfUnset(envSet, "KEY", &flagConfig.Security.Key, *flagKey)
	utils.SetIntIfUnset(envSet, "RATE_LIMIT", &flagConfig.Agent.RateLimit, *flagRateLimit)
	utils.SetStringIfUnset(envSet, "CRYPTO_KEY", &flagConfig.Security.CryptoKey, *flagCryptoKey)

	finalConfig := config.MergeConfigs(flagConfig, configFromFile)

	if finalConfig.Server.Address != "" && finalConfig.Server.Address[:4] != "http" {
		finalConfig.Server.Address = "http://" + finalConfig.Server.Address
	}

	return finalConfig, nil
}
