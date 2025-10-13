package handlers

import (
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/internal/audit"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

// MetricsRepo описывает минимальный набор методов для хранилища, с которым работают HTTP-хендлеры.
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

// Handler хранит экземпляр хранилища метрик, ключ для проверки целостности данных и диспетчер событий аудита.
type Handler struct {
	Repo    MetricsRepo
	Key     string
	Auditor *audit.Publisher
}

func (h *Handler) callAudit(r *http.Request, metricNames []string) {
	if h.Auditor == nil || len(metricNames) == 0 {
		return
	}
	ev := audit.BuildEvent(r, metricNames)
	h.Auditor.Publish(ev)
}
