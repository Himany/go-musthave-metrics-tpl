package agent

import (
	"math/rand/v2"
	"runtime"
	"time"
)

func (a *Agent) collector() {
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
