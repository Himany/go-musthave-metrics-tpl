package agent

import (
	"sync"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
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
	Key            string
	RateLimit      int
	Tasks          chan []models.Metrics
}

func createAgent(url string, reportInterval int, pollInterval int, key string, rateLimit int) *agent {
	return (&agent{
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
	agent := createAgent(cfg.Address, cfg.ReportInterval, cfg.PollInterval, cfg.Key, cfg.RateLimit)

	agent.CreateWorkers()

	go agent.collector()
	go agent.collectorAdv()
	go agent.reportHandler()

	select {}
}
