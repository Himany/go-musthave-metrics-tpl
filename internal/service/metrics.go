package service

import (
	"context"

	apperrors "github.com/Himany/go-musthave-metrics-tpl/internal/errors"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/repository"
)

// MetricsService предоставляет бизнес-логику для работы с метриками
type MetricsService struct {
	repo repository.MetricsRepo
}

// NewMetricsService создает новый экземпляр сервиса метрик
func NewMetricsService(repo repository.MetricsRepo) *MetricsService {
	return &MetricsService{
		repo: repo,
	}
}

// GetAllMetricsData возвращает все метрики в формате для отображения
func (s *MetricsService) GetAllMetricsData(ctx context.Context) (map[string]string, error) {
	result := make(map[string]string)

	// Получаем gauge метрики
	keysGauge, err := s.repo.GetKeyGauge(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range keysGauge {
		value, exists := s.repo.GetGauge(ctx, key)
		if exists {
			result[key] = models.FormatGaugeValue(value)
		}
	}

	// Получаем counter метрики
	keysCounter, err := s.repo.GetKeyCounter(ctx)
	if err != nil {
		return nil, err
	}

	for _, key := range keysCounter {
		value, exists := s.repo.GetCounter(ctx, key)
		if exists {
			result[key] = models.FormatCounterValue(value)
		}
	}

	return result, nil
}

// PingStorage проверяет доступность хранилища
func (s *MetricsService) PingStorage(ctx context.Context) error {
	return s.repo.Ping(ctx)
}

// GetMetric возвращает значение метрики по типу и имени
func (s *MetricsService) GetMetric(ctx context.Context, metricType, name string) (interface{}, error) {
	switch metricType {
	case "gauge":
		value, exists := s.repo.GetGauge(ctx, name)
		if !exists {
			return nil, apperrors.ErrMetricNotFound
		}
		return value, nil
	case "counter":
		value, exists := s.repo.GetCounter(ctx, name)
		if !exists {
			return nil, apperrors.ErrMetricNotFound
		}
		return value, nil
	default:
		return nil, apperrors.ErrUnknownMetricType
	}
}

// GetMetricJSON возвращает метрику в формате JSON
func (s *MetricsService) GetMetricJSON(ctx context.Context, metric models.Metrics) (*models.Metrics, error) {
	if err := s.validateGetMetricJSON(metric); err != nil {
		return nil, err
	}

	result := metric

	switch metric.MType {
	case "gauge":
		value, exists := s.repo.GetGauge(ctx, metric.ID)
		if !exists {
			return nil, apperrors.ErrMetricNotFound
		}
		result.Value = &value
	case "counter":
		value, exists := s.repo.GetCounter(ctx, metric.ID)
		if !exists {
			return nil, apperrors.ErrMetricNotFound
		}
		result.Delta = &value
	default:
		return nil, apperrors.ErrUnknownMetricType
	}

	return &result, nil
}

// UpdateMetric обновляет метрику
func (s *MetricsService) UpdateMetric(ctx context.Context, metric models.Metrics) error {
	if err := s.validateUpdateMetric(metric); err != nil {
		return err
	}

	switch metric.MType {
	case "gauge":
		if metric.Value == nil {
			return apperrors.ErrGaugeValueRequired
		}
		s.repo.UpdateGauge(ctx, metric.ID, *metric.Value)
	case "counter":
		if metric.Delta == nil {
			return apperrors.ErrCounterDeltaRequired
		}
		current, _ := s.repo.GetCounter(ctx, metric.ID)
		s.repo.UpdateCounter(ctx, metric.ID, current+*metric.Delta)
	default:
		return apperrors.ErrUnknownMetricType
	}

	return nil
}

// BatchUpdate обновляет множество метрик одной операцией
func (s *MetricsService) BatchUpdate(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return apperrors.ErrEmptyMetrics
	}

	return s.repo.BatchUpdate(ctx, metrics)
}

// validateGetMetricJSON проверяет корректность данных для получения метрики
func (s *MetricsService) validateGetMetricJSON(metric models.Metrics) error {
	if metric.ID == "" {
		return apperrors.ErrMetricIDRequired
	}
	if metric.MType != "gauge" && metric.MType != "counter" {
		return apperrors.ErrInvalidMetricType
	}
	return nil
}

// validateUpdateMetric проверяет корректность данных для обновления метрики
func (s *MetricsService) validateUpdateMetric(metric models.Metrics) error {
	if metric.ID == "" {
		return apperrors.ErrMetricIDRequired
	}
	if metric.MType != "gauge" && metric.MType != "counter" {
		return apperrors.ErrInvalidMetricType
	}
	return nil
}
