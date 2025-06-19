package main

import (
	"log"

	"github.com/Himany/go-musthave-metrics-tpl/internal/agent"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
)

func main() {
	cfg, err := parseConfig()
	if err != nil {
		log.Fatal("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal("failed to initialize logger: " + err.Error())
	}

	agent.Run(cfg)
}
