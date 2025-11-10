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
const defaultLogLevel = "info"
const defaultStoreInterval = 300
const defaultFileStoragePath = "metrics_data"
const defaultRestore = false
const defaultDataBaseDSN = ""
const defaultAuditFile = ""
const defaultAuditURL = ""
const defaultPprofAddr = ""

func parseFlags() (*config.Config, error) {
	envSet := make(envTracker)
	var flagRunAddr = flag.String("a", defaultRunAddr, "address and port to run server")
	var flagLogLevel = flag.String("l", defaultLogLevel, "log level")
	var flagStoreInterval = flag.Int("i", defaultStoreInterval, "time interval in seconds after which the current server readings are saved to disk")
	var flagFileStoragePath = flag.String("f", defaultFileStoragePath, "the path to the file where the current values are saved")
	var flagRestore = flag.Bool("r", defaultRestore, "whether or not to download previously saved values from the specified file at server startup")
	//host=localhost user=postgres password=123321 dbname=metrics sslmode=disable
	var flagDataBaseDSN = flag.String("d", defaultDataBaseDSN, "A string with settings for connecting the postgresql database")
	var flagKey = flag.String("k", "", "key")
	var flagAuditFile = flag.String("audit-file", defaultAuditFile, "audit file")
	var flagAuditURL = flag.String("audit-url", defaultAuditURL, "audit URL")
	var flagPprofAddr = flag.String("pprof-addr", defaultPprofAddr, "enable pprof on the provided address (empty to disable)")
	var flagCryptoKey = flag.String("crypto-key", "", "path to private key file for asymmetric decryption")
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

	configFromFile, err := config.LoadServerConfigFromFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	config.SetDefaultsForServer(configFromFile)

	var envConfig config.Config
	err = env.ParseWithOptions(&envConfig, env.Options{
		OnSet: func(tag string, value any, isDefault bool) {
			envSet[tag] = true
		},
	})
	if err != nil {
		return nil, err
	}

	flagConfig := &config.Config{}

	utils.SetStringIfUnset(envSet, "ADDRESS", &flagConfig.Server.Address, *flagRunAddr)
	utils.SetStringIfUnset(envSet, "LOGLEVEL", &flagConfig.LogLevel, *flagLogLevel)
	utils.SetIntIfUnset(envSet, "STORE_INTERVAL", &flagConfig.Server.StoreInterval, *flagStoreInterval)
	utils.SetStringIfUnset(envSet, "FILE_STORAGE_PATH", &flagConfig.Storage.FileStoragePath, *flagFileStoragePath)
	utils.SetBoolIfUnset(envSet, "RESTORE", &flagConfig.Server.Restore, *flagRestore)
	utils.SetStringIfUnset(envSet, "DATABASE_DSN", &flagConfig.Database.DSN, *flagDataBaseDSN)
	utils.SetStringIfUnset(envSet, "KEY", &flagConfig.Security.Key, *flagKey)
	utils.SetStringIfUnset(envSet, "AUDIT_FILE", &flagConfig.Audit.File, *flagAuditFile)
	utils.SetStringIfUnset(envSet, "AUDIT_URL", &flagConfig.Audit.URL, *flagAuditURL)
	utils.SetStringIfUnset(envSet, "PPROF_ADDR", &flagConfig.Server.PprofAddr, *flagPprofAddr)
	utils.SetStringIfUnset(envSet, "CRYPTO_KEY", &flagConfig.Security.CryptoKey, *flagCryptoKey)

	finalConfig := config.MergeConfigs(flagConfig, configFromFile)

	return finalConfig, nil
}
