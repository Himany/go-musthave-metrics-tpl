package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

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
}

func (h *Handler) getStringValue(metricType string, metricName string) (string, bool) {
	result := ""

	switch metricType {
	case "gauge":
		value, ok := h.Repo.GetGauge(metricName)
		if !ok {
			return "", false
		}
		result = strconv.FormatFloat(value, 'f', -1, 64)
		return result, ok

	case "counter":
		value, ok := h.Repo.GetCounter(metricName)
		if ok {
			result = strconv.FormatInt(value, 10)
		}
		return result, ok

	default:
		return "", false
	}
}

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

func (h *Handler) GetMetricJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

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
		value, ok := h.Repo.GetGauge(metrics.ID)
		if !(ok) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		metrics.Value = &value
	case "counter":
		value, ok := h.Repo.GetCounter(metrics.ID)
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(resp); err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
	}
}

func (h *Handler) GetPing(w http.ResponseWriter, r *http.Request) {
	if err := h.Repo.Ping(); err != nil {
		logger.Log.Error("GetPing", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	list := make([]string, 0)
	keysGauge, err := h.Repo.GetKeyGauge()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	keysCounter, err := h.Repo.GetKeyCounter()
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

func (h *Handler) UpdateHandlerJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

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

func validateUpdateJSON(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !(metrics.MType == "gauge" || metrics.MType == "counter") {
		return errors.New("field 'type' must have a value of 'gauge' or 'counter'")
	}

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return errors.New("field 'value' is required (with the 'gauge' type)")
		}
	case "counter":
		if metrics.Delta == nil {
			return errors.New("field 'delta' is required (with the 'counter' type)")
		}
	}

	return nil
}

func validateGetMetricJSON(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !(metrics.MType == "gauge" || metrics.MType == "counter") {
		return errors.New("field 'type' must have a value of 'gauge' or 'counter'")
	}

	return nil
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

func (h *Handler) BatchUpdateJSON(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	var metrics []models.Metrics
	var buf bytes.Buffer

	//читаем тело запроса
	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		logger.Log.Error("BatchUpdateJSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//десериализуем JSON в Visitor
	if err = json.Unmarshal(buf.Bytes(), &metrics); err != nil {
		logger.Log.Error("BatchUpdateJSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(metrics) == 0 {
		logger.Log.Error("BatchUpdateJSON", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Обновляем данные
	err = h.Repo.BatchUpdate(metrics)
	if err != nil {
		logger.Log.Error("BatchUpdateJSON", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Отвечаем на запрос
	w.WriteHeader(http.StatusOK)
}
