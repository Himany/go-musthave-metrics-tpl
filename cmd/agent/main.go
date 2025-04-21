package main

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"
)

var (
	PollCount = 0
	url       = "http://localhost:8080/update"
	metrics   = make(map[string]float64)
	mu        sync.Mutex
)

func createRequest(metricType string, name string, value string) {
	resp, err := http.Post(url+"/"+metricType+"/"+name+"/"+value, "text/plain", bytes.NewBuffer(nil))
	if err != nil {
		fmt.Printf("Ошибка (%s): %v\n", url+"/"+metricType+"/"+name+"/"+value, err)
		return
	}
	fmt.Printf("%s: %s\n", url+"/"+metricType+"/"+name+"/"+value, resp.Status)
	resp.Body.Close()
}

func metricHandler(pollInterval int) {
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

		time.Sleep(time.Duration(pollInterval) * time.Second)
	}
}

func reportHandler(reportInterval int) {
	for {
		mu.Lock()
		for key, value := range metrics {
			createRequest("gauge", key, strconv.FormatFloat(value, 'f', -1, 64))
		}
		createRequest("counter", "PollCount", strconv.Itoa(PollCount))
		mu.Unlock()

		time.Sleep(time.Duration(reportInterval) * time.Second)
	}
}

func main() {
	pollInterval := 2
	reportInterval := 10

	go metricHandler(pollInterval)
	go reportHandler(reportInterval)

	select {}
}
