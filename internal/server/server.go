package server

import (
	"context"
	"database/sql"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/crypto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) error {
	var repo handlers.MetricsRepo
	var memStorage *storage.MemStorageData
	var db *sql.DB

	if cfg.Database.DSN != "" {
		var err error
		db, err = sql.Open("pgx", cfg.Database.DSN)
		if err != nil {
			logger.Log.Error("failed to open database (postgres)", zap.Error(err))
		} else {
			repo, err = storage.NewPostgresStorage(db)
			if err != nil {
				logger.Log.Fatal("failed to init database storage", zap.Error(err))
			}
		}
	} else {
		withSync := cfg.Server.StoreInterval == 0 && cfg.Storage.FileStoragePath != ""
		memStorage = storage.NewMemStorage(cfg.Storage.FileStoragePath, withSync)

		if cfg.Storage.FileStoragePath != "" {
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

	handler := &handlers.Handler{
		Storage: handlers.StorageHandler{Repo: repo},
		Signer:  handlers.Signer{Key: cfg.Security.Key},
	}

	startPprof(cfg.Server.PprofAddr)

	decryptor, err := crypto.NewRSAEncryptorFromPrivateKey(cfg.Security.CryptoKey)
	if err != nil {
		return err
	}

	r := CreateRouter(handler, cfg.Security.Key, decryptor, cfg.Server.TrustedSubnet)

	server := &http.Server{
		Addr:    cfg.Server.Address,
		Handler: r,
	}

	go func() {
		logger.Log.Info("Starting HTTP server", zap.String("address", cfg.Server.Address))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server startup failed", zap.Error(err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	<-ctx.Done()

	stop()

	logger.Log.Info("Received shutdown signal, starting graceful shutdown...")

	// Выполняем graceful shutdown
	return gracefulShutdown(server, memStorage, db)
}

// gracefulShutdown выполняет корректное завершение работы сервера
func gracefulShutdown(server *http.Server, memStorage *storage.MemStorageData, db *sql.DB) error {
	logger.Log.Info("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Error("Server shutdown failed", zap.Error(err))
		return err
	}
	logger.Log.Info("HTTP server stopped")

	if memStorage != nil {
		if err := memStorage.SaveData(); err != nil {
			logger.Log.Error("Failed to save data on shutdown", zap.Error(err))
			return err
		}
		logger.Log.Info("Memory storage data saved")
	}

	if db != nil {
		if err := db.Close(); err != nil {
			logger.Log.Error("Failed to close database connection", zap.Error(err))
			return err
		}
		logger.Log.Info("Database connection closed")
	}

	logger.Log.Info("Graceful shutdown completed")
	return nil
}
