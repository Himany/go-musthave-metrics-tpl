package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

// UpdateHandlerQuery принимает значения метрики через параметры URL и обновляет хранилище.
func (h *Handler) UpdateHandlerQuery(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	metricValue := chi.URLParam(r, "value")

	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if metricValue == "" || metricType == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := h.updateDataQuery(r.Context(), metricType, metricName, metricValue); err != nil {
		logger.Log.Error("UpdateHandlerQuery", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	h.Audit.Publish(r, []string{metricName})

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateDataQuery(ctx context.Context, metricType, metricName, metricValue string) error {
	var metric models.Metrics
	metric.ID = metricName
	metric.MType = metricType

	switch metricType {
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("error parsing float: %w", err)
		}
		metric.Value = &val

	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing int: %w", err)
		}
		metric.Delta = &val

	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}

	return h.Service.UpdateMetric(ctx, metric)
}

// UpdateHandlerJSON читает метрику из JSON-запроса и сохраняет её в хранилище.
func (h *Handler) UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics
	var buf bytes.Buffer

	//читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//десериализуем JSON в Visitor
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Обновляем данные через сервис
	if err := h.Service.UpdateMetric(r.Context(), metrics); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получаем обновленные данные
	result, err := h.Service.GetMetricJSON(r.Context(), metrics)
	if err != nil {
		logger.Log.Error("UpdateHandlerJson -> GetMetricJSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Отвечаем на запрос
	resp, err := json.Marshal(result)
	if err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash := bodySignature(resp, h.Signer.Key)
	if hash != nil {
		w.Header().Set("HashSHA256", hex.EncodeToString(hash))
	}

	h.Audit.Publish(r, []string{metrics.ID})

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(resp); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
	}
}
