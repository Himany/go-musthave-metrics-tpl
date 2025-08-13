package storage

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/retry"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type dbStorageData struct {
	db *sql.DB
}

const (
	createGaugeTable = `
	CREATE TABLE IF NOT EXISTS gauges (
		id TEXT PRIMARY KEY,
		value DOUBLE PRECISION NOT NULL
	);`

	createCounterTable = `
	CREATE TABLE IF NOT EXISTS counters (
		id TEXT PRIMARY KEY,
		delta BIGINT NOT NULL
	);`
)

func NewPostgresStorage(db *sql.DB) (*dbStorageData, error) {
	if _, err := db.Exec(createGaugeTable); err != nil {
		return nil, fmt.Errorf("failed to create gauges table: %w", err)
	}
	if _, err := db.Exec(createCounterTable); err != nil {
		return nil, fmt.Errorf("failed to create counters table: %w", err)
	}

	return &dbStorageData{db: db}, nil
}

func (s *dbStorageData) Ping() error {
	err := s.db.Ping()
	return err
}

func (s *dbStorageData) UpdateGauge(name string, value float64) {
	retry.WithRetry(func() error {
		_, err := s.db.Exec(`
			INSERT INTO gauges (id, value) VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE SET value = $2;
		`, name, value)
		return err
	}, isRetriableDBError, "UpdateGauge")
}

func (s *dbStorageData) UpdateCounter(name string, value int64) {
	retry.WithRetry(func() error {
		_, err := s.db.Exec(`
			INSERT INTO counters (id, delta) VALUES ($1, $2)
			ON CONFLICT (id) DO UPDATE SET delta = $2;
		`, name, value)
		return err
	}, isRetriableDBError, "UpdateCounter")
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

func (s *dbStorageData) GetKeyGauge() ([]string, error) {
	rows, err := s.db.Query(`SELECT id FROM gauges`)
	if err != nil {
		logger.Log.Error("DB GetKeyGauge query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)

		if err != nil {
			logger.Log.Error("DB GetKeyGauge scan error", zap.Error(err))
			return nil, err
		}

		keys = append(keys, id)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("DB GetKeyGauge rows error", zap.Error(err))
		return nil, err
	}

	return keys, nil
}

func (s *dbStorageData) GetKeyCounter() ([]string, error) {
	rows, err := s.db.Query(`SELECT id FROM counters`)
	if err != nil {
		logger.Log.Error("DB GetKeyCounter query failed", zap.Error(err))
		return nil, err
	}
	defer rows.Close()

	var keys []string
	for rows.Next() {
		var id string
		err := rows.Scan(&id)

		if err != nil {
			logger.Log.Error("DB GetKeyCounter scan error", zap.Error(err))
			return nil, err
		}

		keys = append(keys, id)
	}

	err = rows.Err()
	if err != nil {
		logger.Log.Error("DB GetKeyCounter rows error", zap.Error(err))
		return nil, err
	}

	return keys, nil
}

func (s *dbStorageData) BatchUpdate(metrics []models.Metrics) error {
	return retry.WithRetry(func() error {
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
	}, isRetriableDBError, "BatchUpdate")
}

func isRetriableDBError(err error) bool {
	var pgErr *pgconn.PgError
	var mapErrors = map[string]bool{
		pgerrcode.ConnectionException:                           true,
		pgerrcode.ConnectionDoesNotExist:                        true,
		pgerrcode.ConnectionFailure:                             true,
		pgerrcode.SQLClientUnableToEstablishSQLConnection:       true,
		pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection: true,
		pgerrcode.TransactionResolutionUnknown:                  true,
	}
	if errors.As(err, &pgErr) {
		if result, ok := mapErrors[pgErr.Code]; ok && result {
			return true
		}
	}
	return false
}
