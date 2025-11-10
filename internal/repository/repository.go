package repository

import (
	"context"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

// MetricsRepo описывает минимальный набор методов для хранилища метрик
type MetricsRepo interface {
	Ping(ctx context.Context) error
	UpdateGauge(ctx context.Context, name string, value float64)
	UpdateCounter(ctx context.Context, name string, value int64)
	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)
	GetKeyGauge(ctx context.Context) ([]string, error)
	GetKeyCounter(ctx context.Context) ([]string, error)
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
}
