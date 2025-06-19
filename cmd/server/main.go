package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
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
