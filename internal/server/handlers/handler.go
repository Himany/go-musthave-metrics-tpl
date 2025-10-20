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

// StorageHandler инкапсулирует доступ к хранилищу метрик.
type StorageHandler struct {
	Repo MetricsRepo
}

// Signer хранит ключ подписи для проверки целостности данных.
type Signer struct {
	Key string
}

// AuditNotifier отвечает за публикацию событий аудита.
type AuditNotifier struct {
	Publisher *audit.Publisher
}

// Publish отправляет событие аудита, если зарегистрированы подписчики.
func (a AuditNotifier) Publish(r *http.Request, metricNames []string) {
	if a.Publisher == nil || len(metricNames) == 0 {
		return
	}
	ev := audit.BuildEvent(r, metricNames)
	a.Publisher.Publish(ev)
}

// Handler объединяет специализированные компоненты для работы HTTP-хендлеров.
type Handler struct {
	Storage StorageHandler
	Signer  Signer
	Audit   AuditNotifier
}
