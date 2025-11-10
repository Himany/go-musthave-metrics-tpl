package main

import (
	"fmt"
	"log"

	"github.com/Himany/go-musthave-metrics-tpl/internal/agent"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
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

	cfg, err := parseConfig()
	if err != nil {
		log.Fatal("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal("failed to initialize logger: " + err.Error())
	}

	agent.Run(cfg)
}
