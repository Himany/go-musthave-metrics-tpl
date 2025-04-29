package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
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

func (a *agent) createRequest(metricType, name, value string) {
	resp, err := a.Client.R().
		SetHeader("Content-Type", "text/plain").
		SetPathParams(map[string]string{
			"mType":  metricType,
			"mName":  name,
			"mValue": value,
		}).
		Post(a.URL + "/{mType}/{mName}/{mValue}")

	if err != nil {
		fmt.Printf("Ошибка (%s): %v\n", a.URL+"/"+metricType+"/"+name+"/"+value, err)
		return
	}
	fmt.Printf("%s: %d\n", a.URL+"/"+metricType+"/"+name+"/"+value, resp.StatusCode())
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
		a.Metrics["RandomValue"] = rand.Float64()

		a.PollCount++
		a.Mutex.Unlock()

		time.Sleep(time.Duration(a.PollInterval) * time.Second)
	}
}

func (a *agent) reportHandler() {
	for {
		a.Mutex.Lock()
		for key, value := range a.Metrics {
			a.createRequest("gauge", key, strconv.FormatFloat(value, 'f', -1, 64))
		}
		a.Mutex.Unlock()
		a.createRequest("counter", "PollCount", strconv.FormatInt(a.PollCount, 10))

		time.Sleep(time.Duration(a.ReportInterval) * time.Second)
	}
}

func main() {
	url, reportInterval, pollInterval, err := parseConfig()
	if err != nil {
		log.Fatal(err)
	}

	agent := createAgent(url, reportInterval, pollInterval)

	go agent.metricHandler()
	go agent.reportHandler()

	select {}
}
