package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	version := buildVersion
	if version == "" {
		version = "N/A"
	}
	date := buildDate
	if date == "" {
		date = "N/A"
	}
	commit := buildCommit
	if commit == "" {
		commit = "N/A"
	}
	fmt.Printf("Build version: %s\n", version)
	fmt.Printf("Build date: %s\n", date)
	fmt.Printf("Build commit: %s\n", commit)

	cfg, err := parseFlags()
	if err != nil {
		log.Fatal("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal("failed to initialize logger: " + err.Error())
	}

	logger.Log.Info("flags", zap.Object("config", cfg))

	if err := server.Run(cfg); err != nil {
		logger.Log.Fatal("main", zap.Error(err))
	}
}
