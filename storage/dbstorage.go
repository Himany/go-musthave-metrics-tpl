package storage

import (
	"database/sql"
	"fmt"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"go.uber.org/zap"
)

type DBStorage interface {
	Ping() error
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetKeyGauge() []string
	GetKeyCounter() []string
	BatchUpdate(metrics []models.Metrics) error
}

type dbStorageData struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) (*dbStorageData, error) {
	const gaugeSchema = `
	CREATE TABLE IF NOT EXISTS gauges (
		id TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);`

	const counterSchema = `
	CREATE TABLE IF NOT EXISTS counters (
		id TEXT PRIMARY KEY,
		delta BIGINT NOT NULL
	);`

	if _, err := db.Exec(gaugeSchema); err != nil {
		return nil, fmt.Errorf("failed to create gauges table: %w", err)
	}
	if _, err := db.Exec(counterSchema); err != nil {
		return nil, fmt.Errorf("failed to create counters table: %w", err)
	}

	return &dbStorageData{db: db}, nil
}

func (s *dbStorageData) Ping() error {
	err := s.db.Ping()
	return err
}

func (s *dbStorageData) UpdateGauge(name string, value float64) {
	_, err := s.db.Exec(`
		INSERT INTO gauges (id, value) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET value = $2;
	`, name, value)
	if err != nil {
		logger.Log.Error("DB UpdateGauge failed", zap.Error(err))
	}
}

func (s *dbStorageData) UpdateCounter(name string, value int64) {
	_, err := s.db.Exec(`
		INSERT INTO counters (id, delta) VALUES ($1, $2)
		ON CONFLICT (id) DO UPDATE SET delta = counters.delta + $2;
	`, name, value)
	if err != nil {
		logger.Log.Error("DB UpdateCounter failed", zap.Error(err))
	}
}

func (s *dbStorageData) GetGauge(name string) (float64, bool) {
	var value float64
	err := s.db.QueryRow(`SELECT value FROM gauges WHERE id = $1`, name).Scan(&value)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("DB GetGauge query failed", zap.Error(err))
		}
		return 0, false
	}
	return value, true
}

func (s *dbStorageData) GetCounter(name string) (int64, bool) {
	var delta int64
	err := s.db.QueryRow(`SELECT delta FROM counters WHERE id = $1`, name).Scan(&delta)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Log.Error("DB GetCounter query failed", zap.Error(err))
		}
		return 0, false
	}
	return delta, true
}

func (s *dbStorageData) GetKeyGauge() []string {
	rows, err := s.db.Query(`SELECT id FROM gauges`)
	if err != nil {
		logger.Log.Error("DB GetKeyGauge query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)

		if err != nil {
			logger.Log.Error("DB GetKeyGauge scan error", zap.Error(err))
			return nil
		}

		keys = append(keys, id)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("DB GetKeyGauge rows error", zap.Error(err))
		return nil
	}

	return keys
}

func (s *dbStorageData) GetKeyCounter() []string {
	rows, err := s.db.Query(`SELECT id FROM counters`)
	if err != nil {
		logger.Log.Error("DB GetKeyCounter query failed", zap.Error(err))
		return nil
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)

		if err != nil {
			logger.Log.Error("DB GetKeyCounter scan error", zap.Error(err))
			return nil
		}

		keys = append(keys, id)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("DB GetKeyCounter rows error", zap.Error(err))
		return nil
	}

	return keys
}

func (s *dbStorageData) BatchUpdate(metrics []models.Metrics) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	defer tx.Rollback()

	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value == nil {
				continue
			}
			_, err := tx.Exec(`
				INSERT INTO gauges (id, value) VALUES ($1, $2)
				ON CONFLICT (id) DO UPDATE SET value = $2;
			`, m.ID, *m.Value)
			if err != nil {
				return err
			}

		case "counter":
			if m.Delta == nil {
				continue
			}
			_, err := tx.Exec(`
				INSERT INTO counters (id, delta) VALUES ($1, $2)
				ON CONFLICT (id) DO UPDATE SET delta = counters.delta + $2;
			`, m.ID, *m.Delta)
			if err != nil {
				return err
			}
		default:
			logger.Log.Warn("BatchUpdate unknown metric type", zap.String("type", m.MType))
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
