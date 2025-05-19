package main

import (
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/server"
	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

func main() {
	if err := parseFlags(); err != nil {
		panic("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(logLevel); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	logger.Log.Info("flags",
		zap.String("runAddr", runAddr),
		zap.String("fileStoragePath", fileStoragePath),
		zap.Int("storeInterval", storeInterval),
		zap.Bool("restore", restore),
	)
	memStorage := storage.NewMemStorage(fileStoragePath, (storeInterval == 0))
	handler := &handlers.Handler{Repo: memStorage}
	if restore && fileStoragePath != "" {
		if err := memStorage.LoadData(); err != nil {
			panic("failed to load data: " + err.Error())
		}
	}
	defer func() {
		if err := memStorage.SaveData(); err != nil {
			logger.Log.Error("failed to save data on shutdown", zap.Error(err))
		}
	}()

	if storeInterval != 0 {
		go memStorage.SaveHandler(storeInterval)
	}

	if err := server.Run(handler, runAddr); err != nil {
		logger.Log.Fatal("main", zap.Error(err))
	}
}
