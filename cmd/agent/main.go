package main

import (
	"fmt"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

type AgentConfig struct {
	URL            string
	PollInterval   int
	ReportInterval int
	Client         *resty.Client
}

var (
	PollCount int64 = 0
	metrics         = make(map[string]float64)
	mu        sync.Mutex
)

func createRequest(cfg *AgentConfig, metricType, name, value string) {
	resp, err := cfg.Client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  metricType,
			"mName":  name,
			"mValue": value,
		}).
		Post(cfg.URL + "/{mType}/{mName}/{mValue}")

	if err != nil {
		fmt.Printf("Ошибка (%s): %v\n", cfg.URL+"/"+metricType+"/"+name+"/"+value, err)
		return
	}
	fmt.Printf("%s: %d\n", cfg.URL+"/"+metricType+"/"+name+"/"+value, resp.StatusCode())
}

func metricHandler(interval int) {
	var s runtime.MemStats

	for {
		runtime.ReadMemStats(&s)

		mu.Lock()
		metrics["Alloc"] = float64(s.Alloc)
		metrics["BuckHashSys"] = float64(s.BuckHashSys)
		metrics["GCCPUFraction"] = s.GCCPUFraction
		metrics["HeapAlloc"] = float64(s.HeapAlloc)
		metrics["HeapIdle"] = float64(s.HeapIdle)
		metrics["HeapInuse"] = float64(s.HeapInuse)
		metrics["HeapObjects"] = float64(s.HeapObjects)
		metrics["HeapReleased"] = float64(s.HeapReleased)
		metrics["HeapSys"] = float64(s.HeapSys)
		metrics["LastGC"] = float64(s.LastGC)
		metrics["Lookups"] = float64(s.Lookups)
		metrics["MCacheInuse"] = float64(s.MCacheInuse)
		metrics["MCacheSys"] = float64(s.MCacheSys)
		metrics["MSpanInuse"] = float64(s.MSpanInuse)
		metrics["MSpanSys"] = float64(s.MSpanSys)
		metrics["Mallocs"] = float64(s.Mallocs)
		metrics["NextGC"] = float64(s.NextGC)
		metrics["NumForcedGC"] = float64(s.NumForcedGC)
		metrics["NumGC"] = float64(s.NumGC)
		metrics["OtherSys"] = float64(s.OtherSys)
		metrics["PauseTotalNs"] = float64(s.PauseTotalNs)
		metrics["StackInuse"] = float64(s.StackInuse)
		metrics["StackSys"] = float64(s.StackSys)
		metrics["Sys"] = float64(s.Sys)
		metrics["TotalAlloc"] = float64(s.TotalAlloc)
		metrics["RandomValue"] = rand.Float64()

		PollCount++
		mu.Unlock()

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func reportHandler(cfg *AgentConfig) {
	for {
		mu.Lock()
		for key, value := range metrics {
			createRequest(cfg, "gauge", key, strconv.FormatFloat(value, 'f', -1, 64))
		}
		mu.Unlock()
		createRequest(cfg, "counter", "PollCount", strconv.FormatInt(PollCount, 10))

		time.Sleep(time.Duration(cfg.ReportInterval) * time.Second)
	}
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		fmt.Printf("ошибка инициализации: %v\n", err)
		return
	}

	go metricHandler(cfg.PollInterval)
	go reportHandler(cfg)

	select {}
}
