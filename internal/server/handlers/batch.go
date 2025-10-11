package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"go.uber.org/zap"
)

func (h *Handler) BatchUpdateJSON(w http.ResponseWriter, r *http.Request) {
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

	var names []string
	for _, m := range metrics {
		names = append(names, m.ID)
	}
	h.callAudit(r, names)

	//Отвечаем на запрос
	w.WriteHeader(http.StatusOK)
}
