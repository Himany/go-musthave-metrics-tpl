package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
)

func ExampleHandler_UpdateHandlerJSON() {
	repo := storage.NewMemStorage("", false)
	handler := &Handler{Repo: repo}

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
	handler := &Handler{Repo: repo}
	repo.UpdateGauge("Alloc", 42)

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
	handler := &Handler{Repo: repo}

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

	gauge, _ := repo.GetGauge("Alloc")
	counter, _ := repo.GetCounter("PollCount")

	fmt.Println(rec.Code)
	fmt.Printf("%.0f %d\n", gauge, counter)
	// Output:
	// 200
	// 42 3
}
