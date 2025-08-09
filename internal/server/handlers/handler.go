package handlers

import "github.com/Himany/go-musthave-metrics-tpl/internal/models"

type MetricsRepo interface {
	Ping() error
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetKeyGauge() ([]string, error)
	GetKeyCounter() ([]string, error)
	BatchUpdate(metrics []models.Metrics) error
}

type Handler struct {
	Repo MetricsRepo
	Key  string
}
