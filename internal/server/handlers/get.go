package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	apperrors "github.com/Himany/go-musthave-metrics-tpl/internal/errors"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetAllMetrics возвращает список всех доступных метрик в виде HTML-страницы.
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	metricsData, err := h.Service.GetAllMetricsData(r.Context())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	list := make([]string, 0, len(metricsData))
	for name, value := range metricsData {
		list = append(list, name+": "+value+";")
	}

	resultString := strings.Join(list, "\n")

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write([]byte(resultString)); err != nil {
		logger.Log.Error("GetAllMetrics", zap.Error(err))
		return
	}
}

// GetPing проверяет подключение к хранилищу и возвращает 200 при успехе.
func (h *Handler) GetPing(w http.ResponseWriter, r *http.Request) {
	if err := h.Service.PingStorage(r.Context()); err != nil {
		logger.Log.Error("GetPing", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

// GetMetricQuery возвращает значение метрики из хранилища.
func (h *Handler) GetMetricQuery(w http.ResponseWriter, r *http.Request) {
	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	if metricName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, err := h.Service.GetMetric(r.Context(), metricType, metricName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var valueStr string
	switch metricType {
	case "gauge":
		if v, ok := value.(float64); ok {
			valueStr = models.FormatGaugeValue(v)
		}
	case "counter":
		if v, ok := value.(int64); ok {
			valueStr = models.FormatCounterValue(v)
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(valueStr)); err != nil {
		logger.Log.Error("GetMetric", zap.Error(err))
	}
}

// GetMetricJSON принимает JSON с описанием метрики и возвращает её текущее значениее.
func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	var metrics models.Metrics
	var buf bytes.Buffer

	//читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//десериализуем JSON в Visitor
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получаем данные через сервис
	result, err := h.Service.GetMetricJSON(r.Context(), metrics)
	if err != nil {
		if errors.Is(err, apperrors.ErrMetricNotFound) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			logger.Log.Error("GetMetricJson", zap.Error(err))
			w.WriteHeader(http.StatusBadRequest)
		}
		return
	}

	//Отвечаем на запрос
	resp, err := json.Marshal(result)
	if err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hash := bodySignature(resp, h.Signer.Key)
	if hash != nil {
		w.Header().Set("HashSHA256", hex.EncodeToString(hash))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(resp); err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
	}
}
