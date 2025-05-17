package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
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

func (h *Handler) GetMetric(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := h.updateData(metricType, metricName, metricValue); err != nil {
		logger.Log.Error("GetMetric", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) updateData(metricType, metricName, metricValue string) error {
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
