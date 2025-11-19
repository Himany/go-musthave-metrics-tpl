package agent

import (
	"context"
	"sync"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/config"
	"github.com/Himany/go-musthave-metrics-tpl/internal/crypto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

// Agent собирает системные метрики и отправляет их на сервер.
type Agent struct {
	URL            string
	ReportInterval int
	PollInterval   int
	Client         *resty.Client
	PollCount      int64
	Metrics        map[string]float64
	mutex          sync.Mutex
	Key            string
	RateLimit      int
	Tasks          chan []models.Metrics
	Encryptor      *crypto.RSAEncryptor

	// Поля для graceful shutdown
	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func CreateAgent(cfg *config.Config) (*Agent, error) {
	encryptor, err := crypto.NewRSAEncryptorFromPublicKey(cfg.Security.CryptoKey)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		URL:            cfg.Server.Address,
		ReportInterval: cfg.Agent.ReportInterval,
		PollInterval:   cfg.Agent.PollInterval,
		Client:         resty.New(),
		PollCount:      0,
		Metrics:        make(map[string]float64),
		Key:            cfg.Security.Key,
		RateLimit:      cfg.Agent.RateLimit,
		Tasks:          make(chan []models.Metrics, cfg.Agent.RateLimit*2),
		Encryptor:      encryptor,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

func (a *Agent) Start() error {
	a.CreateWorkers()

	a.wg.Add(3)
	go a.collector()
	go a.collectorAdv()
	go a.reportHandler()

	return nil
}

func (a *Agent) Stop() {
	logger.Log.Info("Stopping agent...")

	a.cancel()

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("All agent goroutines stopped")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("Timeout waiting for agent goroutines to stop")
	}

	a.sendFinalMetrics()

	logger.Log.Info("Agent stopped")
}

func (a *Agent) sendFinalMetrics() {
	logger.Log.Info("Sending final metrics...")

	a.mutex.Lock()
	defer a.mutex.Unlock()

	if len(a.Metrics) == 0 && a.PollCount == 0 {
		logger.Log.Info("No metrics to send")
		return
	}

	var batch []models.Metrics
	for key, value := range a.Metrics {
		val := value
		batch = append(batch, models.Metrics{
			ID:    key,
			MType: "gauge",
			Value: &val,
		})
	}

	if a.PollCount > 0 {
		batch = append(batch, models.Metrics{
			ID:    "PollCount",
			MType: "counter",
			Delta: &a.PollCount,
		})
	}

	if len(batch) > 0 {
		err := a.createBatchRequest(batch)
		if err != nil {
			logger.Log.Error("Failed to send final metrics", zap.Error(err))
		} else {
			logger.Log.Info("Final metrics sent successfully", zap.Int("count", len(batch)))
		}
	}
}
