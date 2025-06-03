package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"math/rand/v2"
	"runtime"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

type agent struct {
	URL            string
	ReportInterval int
	PollInterval   int
	Client         *resty.Client
	PollCount      int64
	Metrics        map[string]float64
	Mutex          sync.Mutex
}

func createAgent(url string, reportInterval int, pollInterval int) *agent {
	return (&agent{
		URL:            url,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		Client:         resty.New(),
		PollCount:      0,
		Metrics:        make(map[string]float64),
	})
}

func compressBody(v models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func compressBatchBody(v []models.Metrics) ([]byte, error) {
	jsonData, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(jsonData); err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (a *agent) createBatchRequest(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return errors.New("empty metrics")
	}

	body, err := compressBatchBody(metrics)
	if err != nil {
		return err
	}

	start := time.Now()

	resp, err := a.Client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(a.URL + "/updates/")

	if err != nil {
		return err
	}

	duration := time.Since(start)

	logger.Log.Info("HTTP BATCH request",
		zap.String("uri", a.URL+"/updates/"),
		zap.String("method", "POST"),
		zap.Duration("duration", duration),
	)
	logger.Log.Info("HTTP BATCH answer",
		zap.Int("status", resp.StatusCode()),
		zap.Int("size", len(resp.Body())),
		zap.String("body", resp.String()),
	)

	return nil
}

func (a *agent) createRequest(metricType string, name string, delta *int64, value *float64) {
	metrics := models.Metrics{
		ID:    name,
		MType: metricType,
		Delta: delta,
		Value: value,
	}

	body, err := compressBody(metrics)
	if err != nil {
		logger.Log.Error("compressBody", zap.Error(err))
		return
	}

	start := time.Now()

	resp, err := a.Client.R().
		SetHeader("Content-Encoding", "gzip").
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(a.URL + "/update/")

	if err != nil {
		logger.Log.Error("createRequest", zap.Error(err))
		return
	}

	duration := time.Since(start)

	logger.Log.Info("HTTP request",
		zap.String("uri", a.URL+"/update/"),
		zap.String("method", "POST"),
		zap.Duration("duration", duration),
	)
	logger.Log.Info("HTTP answer",
		zap.Int("status", resp.StatusCode()),
		zap.Int("size", len(resp.Body())),
		zap.String("body", resp.String()),
	)
}

func (a *agent) metricHandler() {
	var s runtime.MemStats

	for {
		runtime.ReadMemStats(&s)

		a.Mutex.Lock()
		a.Metrics["Alloc"] = float64(s.Alloc)
		a.Metrics["BuckHashSys"] = float64(s.BuckHashSys)
		a.Metrics["GCCPUFraction"] = s.GCCPUFraction
		a.Metrics["HeapAlloc"] = float64(s.HeapAlloc)
		a.Metrics["HeapIdle"] = float64(s.HeapIdle)
		a.Metrics["HeapInuse"] = float64(s.HeapInuse)
		a.Metrics["HeapObjects"] = float64(s.HeapObjects)
		a.Metrics["HeapReleased"] = float64(s.HeapReleased)
		a.Metrics["HeapSys"] = float64(s.HeapSys)
		a.Metrics["LastGC"] = float64(s.LastGC)
		a.Metrics["Lookups"] = float64(s.Lookups)
		a.Metrics["MCacheInuse"] = float64(s.MCacheInuse)
		a.Metrics["MCacheSys"] = float64(s.MCacheSys)
		a.Metrics["MSpanInuse"] = float64(s.MSpanInuse)
		a.Metrics["MSpanSys"] = float64(s.MSpanSys)
		a.Metrics["Mallocs"] = float64(s.Mallocs)
		a.Metrics["NextGC"] = float64(s.NextGC)
		a.Metrics["NumForcedGC"] = float64(s.NumForcedGC)
		a.Metrics["NumGC"] = float64(s.NumGC)
		a.Metrics["OtherSys"] = float64(s.OtherSys)
		a.Metrics["PauseTotalNs"] = float64(s.PauseTotalNs)
		a.Metrics["StackInuse"] = float64(s.StackInuse)
		a.Metrics["StackSys"] = float64(s.StackSys)
		a.Metrics["Sys"] = float64(s.Sys)
		a.Metrics["TotalAlloc"] = float64(s.TotalAlloc)
		a.Metrics["Frees"] = float64(s.Frees)
		a.Metrics["GCSys"] = float64(s.GCSys)
		a.Metrics["RandomValue"] = rand.Float64()

		a.PollCount++
		a.Mutex.Unlock()

		time.Sleep(time.Duration(a.PollInterval) * time.Second)
	}
}

func (a *agent) reportHandler() {
	for {
		/*
			a.Mutex.Lock()
			for key := range a.Metrics {
				value := a.Metrics[key]
				a.createRequest("gauge", key, nil, &value)
			}
			a.Mutex.Unlock()
			a.createRequest("counter", "PollCount", &a.PollCount, nil)
		*/

		a.Mutex.Lock()

		var batch []models.Metrics
		for key, value := range a.Metrics {
			val := value
			batch = append(batch, models.Metrics{
				ID:    key,
				MType: "gauge",
				Value: &val,
			})
		}
		batch = append(batch, models.Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &a.PollCount,
		})

		a.Mutex.Unlock()
		err := a.createBatchRequest(batch)
		if err != nil {
			logger.Log.Error("createBatchRequest", zap.Error(err))
		}

		time.Sleep(time.Duration(a.ReportInterval) * time.Second)
	}
}

func main() {
	url, reportInterval, pollInterval, logLevel, err := parseConfig()
	if err != nil {
		panic("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(logLevel); err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	agent := createAgent(url, reportInterval, pollInterval)

	go agent.metricHandler()
	go agent.reportHandler()

	select {}
}
