package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/service"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
)

func ExampleHandler_UpdateHandlerJSON() {
	repo := storage.NewMemStorage("", false)
	metricsService := service.NewMetricsService(repo)
	handler := &Handler{
		Storage: StorageHandler{Repo: repo},
		Service: metricsService,
	}

	value := 42.0
	metric := models.Metrics{ID: "Alloc", MType: "gauge", Value: &value}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.UpdateHandlerJSON(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":42}
}

func ExampleHandler_GetMetricJSON() {
	repo := storage.NewMemStorage("", false)
	metricsService := service.NewMetricsService(repo)
	handler := &Handler{
		Storage: StorageHandler{Repo: repo},
		Service: metricsService,
	}
	repo.UpdateGauge(context.Background(), "Alloc", 42)

	metric := models.Metrics{ID: "Alloc", MType: "gauge"}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/value/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.GetMetricJSON(rec, req)

	fmt.Println(rec.Code)
	fmt.Println(rec.Body.String())
	// Output:
	// 200
	// {"id":"Alloc","type":"gauge","value":42}
}

func ExampleHandler_BatchUpdateJSON() {
	repo := storage.NewMemStorage("", false)
	metricsService := service.NewMetricsService(repo)
	handler := &Handler{
		Storage: StorageHandler{Repo: repo},
		Service: metricsService,
	}

	value := 42.0
	delta := int64(3)
	metrics := []models.Metrics{
		{ID: "Alloc", MType: "gauge", Value: &value},
		{ID: "PollCount", MType: "counter", Delta: &delta},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates/", bytes.NewReader(body))
	rec := httptest.NewRecorder()

	handler.BatchUpdateJSON(rec, req)

	gauge, _ := repo.GetGauge(context.Background(), "Alloc")
	counter, _ := repo.GetCounter(context.Background(), "PollCount")

	fmt.Println(rec.Code)
	fmt.Printf("%.0f %d\n", gauge, counter)
	// Output:
	// 200
	// 42 3
}
