package agent

import (
	"sync"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
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

func Run(cfg *config.Config) error {
	agent := createAgent(cfg.Address, cfg.ReportInterval, cfg.PollInterval)

	go agent.collector()
	go agent.reportHandler()

	select {}
}
