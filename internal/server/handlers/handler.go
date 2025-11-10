package handlers

import (
	"context"
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/internal/audit"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/repository"
)

// MetricsRepo описывает минимальный набор методов для хранилища, с которым работают HTTP-хендлеры.
type MetricsRepo = repository.MetricsRepo

// MetricsService описывает интерфейс сервиса для работы с метриками
type MetricsService interface {
	GetAllMetricsData(ctx context.Context) (map[string]string, error)
	PingStorage(ctx context.Context) error
	GetMetric(ctx context.Context, metricType, name string) (interface{}, error)
	GetMetricJSON(ctx context.Context, metric models.Metrics) (*models.Metrics, error)
	UpdateMetric(ctx context.Context, metric models.Metrics) error
	BatchUpdate(ctx context.Context, metrics []models.Metrics) error
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
	Storage StorageHandler // Оставляем для обратной совместимости
	Service MetricsService // Новый сервисный слой
	Signer  Signer
	Audit   AuditNotifier
}
