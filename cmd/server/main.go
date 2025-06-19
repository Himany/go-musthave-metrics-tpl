package main

import (
	"database/sql"
	"log"

	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/server"
	"github.com/Himany/go-musthave-metrics-tpl/storage"

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

	var repo handlers.MetricsRepo
	if cfg.DataBaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DataBaseDSN)
		if err != nil {
			logger.Log.Error("failed to open database (postgres)", zap.Error(err))
		} else {
			repo, err = storage.NewPostgresStorage(db)
			if err != nil {
				logger.Log.Fatal("failed to init database storage", zap.Error(err))
			}
			defer db.Close()
		}
	} else {
		withSync := cfg.StoreInterval == 0 && cfg.FileStoragePath != ""
		memStorage := storage.NewMemStorage(cfg.FileStoragePath, withSync)
		if withSync {
			defer func() {
				if err := memStorage.SaveData(); err != nil {
					logger.Log.Error("failed to save data on shutdown", zap.Error(err))
				}
			}()
			if cfg.Restore {
				if err := memStorage.LoadData(); err != nil {
					logger.Log.Fatal("failed to load data", zap.Error(err))
				}
			}
			if cfg.StoreInterval != 0 {
				go memStorage.SaveHandler(cfg.StoreInterval)
			}
		}
		repo = memStorage
	}

	handler := &handlers.Handler{Repo: repo}
	if err := server.Run(handler, cfg.Address); err != nil {
		logger.Log.Fatal("main", zap.Error(err))
	}
}
