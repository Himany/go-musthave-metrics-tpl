package agent

import (
	"sync"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/go-resty/resty/v2"
)

// Agent собирает системные метрики и отправляет их на сервер.
type Agent struct {
	URL            string
	ReportInterval int
	PollInterval   int
	Client         *resty.Client
	PollCount      int64
	Metrics        map[string]float64
	Mutex          sync.Mutex
	Key            string
	RateLimit      int
	Tasks          chan []models.Metrics
}

func createAgent(url string, reportInterval int, pollInterval int, key string, rateLimit int) *Agent {
	return (&Agent{
		URL:            url,
		ReportInterval: reportInterval,
		PollInterval:   pollInterval,
		Client:         resty.New(),
		PollCount:      0,
		Metrics:        make(map[string]float64),
		Key:            key,
		RateLimit:      rateLimit,
		Tasks:          make(chan []models.Metrics, rateLimit*2),
	})
}

func Run(cfg *config.Config) error {
	ag := createAgent(cfg.Server.Address, cfg.Agent.ReportInterval, cfg.Agent.PollInterval, cfg.Security.Key, cfg.Agent.RateLimit)

	ag.CreateWorkers()

	go ag.collector()
	go ag.collectorAdv()
	go ag.reportHandler()

	stop := make(chan struct{})
	<-stop

	return nil
}
