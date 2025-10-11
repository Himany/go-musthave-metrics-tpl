package server

import (
	"database/sql"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/audit"
	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) error {
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
		if cfg.FileStoragePath != "" {
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

	publisher := audit.NewPublisher()
	if cfg.AuditFile != "" {
		publisher.Register(audit.NewFileSink(cfg.AuditFile))
	}
	if cfg.AuditURL != "" {
		publisher.Register(audit.NewHTTPSink(cfg.AuditURL, 5*time.Second))
	}

	handler := &handlers.Handler{
		Repo:    repo,
		Key:     cfg.Key,
		Auditor: publisher,
	}
	if err := Router(handler, cfg.Address, cfg.Key); err != nil {
		return err
	}

	return nil
}
