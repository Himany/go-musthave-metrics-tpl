package handlers

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/storage"
)

func benchHandler(b *testing.B) *Handler {
	b.Helper()
	ms := storage.NewMemStorage("", false)
	return &Handler{Storage: StorageHandler{Repo: ms}}
}

func genOneMetric(i int) []byte {
	v := float64(i%1000) + 0.5
	m := models.Metrics{ID: "gauge_metric", MType: "gauge", Value: &v}
	data, _ := json.Marshal(m)
	return data
}

func genBatchMetrics(n int) []byte {
	arr := make([]models.Metrics, 0, n)
	for i := 0; i < n; i++ {
		if i%2 == 0 {
			v := float64(i%1000) + 0.5
			arr = append(arr, models.Metrics{ID: "gauge_metric", MType: "gauge", Value: &v})
		} else {
			d := int64(i % 1000)
			arr = append(arr, models.Metrics{ID: "counter_metric", MType: "counter", Delta: &d})
		}
	}
	data, _ := json.Marshal(arr)
	return data
}

func Benchmark_UpdateMetricJSON(b *testing.B) {
	h := benchHandler(b)

	b.ReportAllocs()
	b.ReportMetric(1, "records/op")

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		payload := genOneMetric(rng.Int())
		b.StartTimer()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		h.UpdateHandlerJSON(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}

func Benchmark_UpdatesBatch_5k(b *testing.B) {
	h := benchHandler(b)
	const n = 5000

	b.ReportAllocs()
	b.ReportMetric(n, "records/op")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		payload := genBatchMetrics(n)
		b.StartTimer()

		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json; charset=utf-8")

		h.BatchUpdateJSON(rr, req)
		if rr.Code != http.StatusOK {
			b.Fatalf("unexpected status: %d", rr.Code)
		}
	}
}
