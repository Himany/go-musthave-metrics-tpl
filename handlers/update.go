package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

type Handler struct {
	repo storage.Storage
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
	if metricName == "" || metricType == "" {
		if metricName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	value, isOk := getStringValue(h, metricType, metricName)
	if !isOk {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(value))
}

func (h *Handler) GetAllMetrics(w http.ResponseWriter, r *http.Request) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" && !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	list := make([]string, 0)
	for _, item := range h.Repo.GetKeyGauge() {
		value, ok := getStringValue(h, "gauge", item)
		if ok {
			list = append(list, (item + ": " + value + ";"))
		}
	}
	for _, item := range h.Repo.GetKeyCounter() {
		value, ok := getStringValue(h, "counter", item)
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

	if metricValue == "" || metricName == "" || metricType == "" {
		if metricName == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	status := h.updateData(metricType, metricName, metricValue)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
}

func (h *Handler) updateData(metricType, metricName, metricValue string) int {
	switch metricType {
	case "gauge":
		val, err := strconv.ParseFloat(metricValue, 64)
		if err != nil {
			return http.StatusBadRequest
		}
		h.Repo.UpdateGauge(metricName, val)
		return http.StatusOK

	case "counter":
		val, err := strconv.ParseInt(metricValue, 10, 64)
		if err != nil {
			return http.StatusBadRequest
		}
		old, ok := h.Repo.GetCounter(metricName)
		if ok {
			val += old
		}
		h.Repo.UpdateCounter(metricName, val)
		return http.StatusOK

	default:
		return http.StatusBadRequest
	}
}
