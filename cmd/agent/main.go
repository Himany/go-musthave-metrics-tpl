package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

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

	ag, err := agent.CreateAgent(cfg)
	if err != nil {
		log.Fatal("failed to create agent: " + err.Error())
	}

	if err := ag.Start(); err != nil {
		log.Fatal("failed to start agent: " + err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	logger.Log.Info("Agent started successfully")

	<-ctx.Done()

	stop()

	logger.Log.Info("Received shutdown signal, starting graceful shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	shutdownDone := make(chan struct{})
	go func() {
		ag.Stop()
		close(shutdownDone)
	}()

	select {
	case <-shutdownDone:
		logger.Log.Info("Agent stopped gracefully")
	case <-shutdownCtx.Done():
		logger.Log.Warn("Shutdown timeout exceeded, forcing exit")
	}

	logger.Log.Info("Application terminated")
}
