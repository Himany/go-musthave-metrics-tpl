package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

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

	if err := h.updateDataQuery(metricType, metricName, metricValue); err != nil {
		logger.Log.Error("UpdateHandlerQuery", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateDataQuery(metricType, metricName, metricValue string) error {
	switch metricType {
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return fmt.Errorf("error parsing float: %w", err)
		}
		h.Repo.UpdateGauge(metricName, val)
		return nil

	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return fmt.Errorf("error parsing int: %w", err)
		}
		old, ok := h.Repo.GetCounter(metricName)
		if ok {
			val += old
		}
		h.Repo.UpdateCounter(metricName, val)
		return nil

	default:
		return fmt.Errorf("unknown metric type: %s", metricType)
	}
}

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

	//Проверка объекта
	if err = validateUpdateJSON(metrics); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Обновляем данные
	switch metrics.MType {
	case "gauge":
		h.Repo.UpdateGauge(metrics.ID, *metrics.Value) //Обновляем данные в хранилище
		newValue, ok := h.Repo.GetGauge(metrics.ID)    //Обновляем данные в структуре для ответа
		if ok {
			*metrics.Value = newValue
		} else {
			logger.Log.Error("UpdateHandlerJSON -> GetGauge", zap.Error(err))
		}

	case "counter":
		value := *metrics.Delta
		old, ok := h.Repo.GetCounter(metrics.ID)
		if ok {
			value += old
		}
		h.Repo.UpdateCounter(metrics.ID, value)       //Обновляем данные в хранилище
		newValue, ok := h.Repo.GetCounter(metrics.ID) //Обновляем данные в структуре для ответа
		if ok {
			*metrics.Delta = newValue
		} else {
			logger.Log.Error("UpdateHandlerJSON -> GetCounter", zap.Error(err))
		}
	}

	//Отвечаем на запрос
	resp, err := json.Marshal(metrics)
	if err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(resp); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
	}
}
