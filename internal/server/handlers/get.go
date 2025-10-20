package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

// GetAllMetrics возвращает список всех доступных метрик в виде HTML-страницы.
func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	list := make([]string, 0)
	keysGauge, err := h.Storage.Repo.GetKeyGauge()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	keysCounter, err := h.Storage.Repo.GetKeyCounter()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for _, item := range keysGauge {
		value, ok := h.getStringValue("gauge", item)
		if ok {
			list = append(list, (item + ": " + value + ";"))
		}
	}
	for _, item := range keysCounter {
		value, ok := h.getStringValue("counter", item)
		if ok {
			list = append(list, (item + ": " + value + ";"))
		}
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
	if err := h.Storage.Repo.Ping(); err != nil {
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
		w.WriteHeader(http.StatusNotFound)
		return
	}

	value, ok := h.getStringValue(metricType, metricName)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(value)); err != nil {
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

	//Проверка объекта
	if err = validateGetMetricJSON(metrics); err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получаем данные
	switch metrics.MType {
	case "gauge":
		value, ok := h.Storage.Repo.GetGauge(metrics.ID)
		if !(ok) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metrics.Value = &value
	case "counter":
		value, ok := h.Storage.Repo.GetCounter(metrics.ID)
		if !(ok) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metrics.Delta = &value
	}

	//Отвечаем на запрос
	resp, err := json.Marshal(metrics)
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
