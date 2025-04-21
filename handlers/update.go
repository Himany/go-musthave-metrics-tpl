package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

type Handler struct {
	Repo storage.Storage
}

func (h *Handler) UpdateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if !strings.HasPrefix(r.Header.Get("Content-Type"), "text/plain") {
		w.WriteHeader(http.StatusUnsupportedMediaType)
		return
	}

	pathValues := strings.Split(r.URL.Path, "/")
	if len(pathValues) < 5 {
		if len(pathValues) < 4 || pathValues[3] == "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	metricType := pathValues[2]
	metricName := pathValues[3]
	metricValue := pathValues[4]

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
