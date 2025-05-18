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
	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

type Handler struct {
	Repo storage.Storage
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
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	metricType := chi.URLParam(r, "type")
	metricName := chi.URLParam(r, "name")
	if metricName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if metricType == "" {
		w.WriteHeader(http.StatusBadRequest)
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

func (h *Handler) GetMetricJson(w http.ResponseWriter, r *http.Request) {
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
	if err = validateGetMetricJson(metrics); err != nil {
		logger.Log.Error("GetMetricJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Получаем данные
	var ok bool
	switch metrics.MType {
	case "gauge":
		*metrics.Value, ok = h.Repo.GetGauge(metrics.ID)
	case "counter":
		*metrics.Delta, ok = h.Repo.GetCounter(metrics.ID)
	}

	if !(ok) {
		w.WriteHeader(http.StatusNotFound)
		return
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

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	list := make([]string, 0)
	for _, item := range h.Repo.GetKeyGauge() {
		value, ok := h.getStringValue("gauge", item)
		if ok {
			list = append(list, (item + ": " + value + ";"))
		}
	}
	for _, item := range h.Repo.GetKeyCounter() {
		value, ok := h.getStringValue("counter", item)
		if ok {
			list = append(list, (item + ": " + value + ";"))
		}
	}
	resultString := strings.Join(list, "\n")

	w.Header().Set("Content-Type", "text/plain")
	_, _ = w.Write([]byte(resultString))
}

func (h *Handler) UpdateHandlerQuery(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

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

func (h *Handler) UpdateHandlerJson(w http.ResponseWriter, r *http.Request) {
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
	if err = validateUpdateJson(metrics); err != nil {
		logger.Log.Error("UpdateHandlerJson", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Обновляем данные
	switch metrics.MType {
	case "gauge":
		h.Repo.UpdateGauge(metrics.ID, *metrics.Value)  //Обновляем данные в хранилище
		*metrics.Value, _ = h.Repo.GetGauge(metrics.ID) //Обновляем данные в структуре для ответа
	case "counter":
		h.Repo.UpdateCounter(metrics.ID, *metrics.Delta)  //Обновляем данные в хранилище
		*metrics.Delta, _ = h.Repo.GetCounter(metrics.ID) //Обновляем данные в структуре для ответа
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

func validateUpdateJson(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !((metrics.MType == "gauge") || (metrics.MType == "counter")) {
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

func validateGetMetricJson(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !((metrics.MType == "gauge") || (metrics.MType == "counter")) {
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
