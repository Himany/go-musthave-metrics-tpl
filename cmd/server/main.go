package main

import (
	"database/sql"

	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/server"
	"github.com/Himany/go-musthave-metrics-tpl/storage"

	_ "github.com/jackc/pgx/v5/stdlib"
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
		zap.String("dataBaseDSN", dataBaseDSN),
	)

	handler := &handlers.Handler{}
	if dataBaseDSN != "" {
		db, err := sql.Open("pgx", dataBaseDSN)
		if err != nil {
			logger.Log.Error("failed to open database (postgres)", zap.Error(err))
		} else {
			repo, err := storage.NewPostgresStorage(db)
			if err != nil {
				db.Close()
				logger.Log.Fatal("failed to init database storage", zap.Error(err))
			}
			handler.Repo = repo
			defer db.Close()
		}
	} else {
		memStorage := storage.NewMemStorage(fileStoragePath, ((storeInterval == 0) && fileStoragePath != ""))
		handler.Repo = memStorage
		if restore && fileStoragePath != "" {
			if err := memStorage.LoadData(); err != nil {
				logger.Log.Fatal("failed to load data", zap.Error(err))
			}
		}
		if fileStoragePath != "" {
			defer func() {
				if err := memStorage.SaveData(); err != nil {
					logger.Log.Error("failed to save data on shutdown", zap.Error(err))
				}
			}()
			if storeInterval != 0 {
				go memStorage.SaveHandler(storeInterval)
			}
		}
	}

	if err := server.Run(handler, runAddr); err != nil {
		logger.Log.Fatal("main", zap.Error(err))
	}
}
