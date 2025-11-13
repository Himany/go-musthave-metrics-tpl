package server

import (
	"database/sql"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/audit"
	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/crypto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/service"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) error {
	var repo handlers.MetricsRepo
	if cfg.Database.DSN != "" {
		db, err := sql.Open("pgx", cfg.Database.DSN)
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
		withSync := cfg.Server.StoreInterval == 0 && cfg.Storage.FileStoragePath != ""
		memStorage := storage.NewMemStorage(cfg.Storage.FileStoragePath, withSync)
		if cfg.Storage.FileStoragePath != "" {
			defer func() {
				if err := memStorage.SaveData(); err != nil {
					logger.Log.Error("failed to save data on shutdown", zap.Error(err))
				}
			}()
			if cfg.Server.Restore {
				if err := memStorage.LoadData(); err != nil {
					logger.Log.Fatal("failed to load data", zap.Error(err))
				}
			}
			if cfg.Server.StoreInterval != 0 {
				go memStorage.SaveHandler(cfg.Server.StoreInterval)
			}
		}
		repo = memStorage
	}

	publisher := audit.NewPublisher()
	if cfg.Audit.File != "" {
		publisher.Register(audit.NewFileSink(cfg.Audit.File))
	}
	if cfg.Audit.URL != "" {
		publisher.Register(audit.NewHTTPSink(cfg.Audit.URL, 5*time.Second))
	}

	// Создаем сервис метрик
	metricsService := service.NewMetricsService(repo)

	handler := &handlers.Handler{
		Storage: handlers.StorageHandler{Repo: repo}, // Для обратной совместимости
		Service: metricsService,                      // Новый сервисный слой
		Signer:  handlers.Signer{Key: cfg.Security.Key},
		Audit:   handlers.AuditNotifier{Publisher: publisher},
	}

	startPprof(cfg.Server.PprofAddr)

	decryptor, err := crypto.NewRSAEncryptorFromPrivateKey(cfg.Security.CryptoKey)
	if err != nil {
		return err
	}

	if err := Router(handler, cfg.Server.Address, cfg.Security.Key, decryptor); err != nil {
		return err
	}

	return nil
}
