package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"log"
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

func (a *agent) retryGzipJSONRequest(body []byte, route string) (*resty.Response, time.Duration, error) {
	retryDelays := []int{0, 1, 3, 5}
	var lastErr error
	var lastResp *resty.Response

	start := time.Now()
	for attempt := 0; attempt < len(retryDelays); attempt++ {
		resp, err := a.Client.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(body).
			Post(a.URL + route)

		if err == nil {
			duration := time.Since(start)

			return resp, duration, nil
		} else {
			lastErr = err
			lastResp = resp
		}

		nonBreakingCodes := map[int]bool{
			502: true,
			503: true,
			504: true,
			429: true,
		}
		if _, nonBreaking := nonBreakingCodes[resp.StatusCode()]; resp != nil && !nonBreaking {
			break
		}

		if retryDelays[attempt] != 0 {
			time.Sleep(time.Duration(retryDelays[attempt]) * time.Second)
		}
	}

	return lastResp, time.Since(start), lastErr
}

func (a *agent) createBatchRequest(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return errors.New("empty metrics")
	}

	body, err := compressBatchBody(metrics)
	if err != nil {
		return err
	}

	var route = "/updates/"
	resp, duration, err := a.retryGzipJSONRequest(body, route)

	if err == nil && resp != nil {
		logger.Log.Info("HTTP BATCH",
			zap.String("uri (request)", a.URL+route),
			zap.String("method (request)", "POST"),
			zap.Duration("duration", duration),
			zap.Int("status (answer)", resp.StatusCode()),
			zap.Int("size (answer)", len(resp.Body())),
			zap.String("body (answer)", resp.String()),
		)
		return nil
	}

	return err
}

func (a *agent) createRequest(metricType string, name string, delta *int64, value *float64) error {
	metrics := models.Metrics{
		ID:    name,
		MType: metricType,
		Delta: delta,
		Value: value,
	}

	body, err := compressBody(metrics)
	if err != nil {
		return err
	}

	var route = "/update/"
	resp, duration, err := a.retryGzipJSONRequest(body, route)

	if err == nil && resp != nil {
		logger.Log.Info("HTTP request",
			zap.String("uri", a.URL+route),
			zap.String("method", "POST"),
			zap.Duration("duration", duration),
		)

		logger.Log.Info("HTTP answer",
			zap.Int("status", resp.StatusCode()),
			zap.Int("size", len(resp.Body())),
			zap.String("body", resp.String()),
		)

		return nil
	}

	return err
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
	cfg, err := parseConfig()
	if err != nil {
		log.Fatal("failed to initialize flags: " + err.Error())
	}

	if err := logger.Initialize(cfg.LogLevel); err != nil {
		log.Fatal("failed to initialize logger: " + err.Error())
	}

	agent := createAgent(cfg.Address, cfg.ReportInterval, cfg.PollInterval)

	go agent.metricHandler()
	go agent.reportHandler()

	select {}
}
