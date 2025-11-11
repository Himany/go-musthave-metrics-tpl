package agent

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	Mutex          sync.Mutex
	Key            string
	RateLimit      int
	Tasks          chan []models.Metrics
	Encryptor      *crypto.RSAEncryptor

	// Поля для graceful shutdown
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
}

func createAgent(url string, reportInterval int, pollInterval int, key string, rateLimit int, encryptor *crypto.RSAEncryptor) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

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
		Encryptor:      encryptor,
		shutdownChan:   make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
	})
}

func Run(cfg *config.Config) error {
	encryptor, err := crypto.NewRSAEncryptorFromPublicKey(cfg.Security.CryptoKey)
	if err != nil {
		return err
	}

	ag := createAgent(cfg.Server.Address, cfg.Agent.ReportInterval, cfg.Agent.PollInterval, cfg.Security.Key, cfg.Agent.RateLimit, encryptor)

	ag.CreateWorkers()

	ag.wg.Add(3)
	go ag.collector()
	go ag.collectorAdv()
	go ag.reportHandler()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	logger.Log.Info("Agent started successfully")

	select {
	case sig := <-sigChan:
		logger.Log.Info("Received shutdown signal", zap.String("signal", sig.String()))
		ag.gracefulShutdown()
	case <-ag.shutdownChan:
		logger.Log.Info("Shutdown requested")
	}

	logger.Log.Info("Agent stopped")
	return nil
}

// gracefulShutdown выполняет корректное завершение работы агента
func (a *Agent) gracefulShutdown() {
	logger.Log.Info("Starting graceful shutdown...")

	a.cancel()

	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logger.Log.Info("All goroutines stopped")
	case <-time.After(10 * time.Second):
		logger.Log.Warn("Timeout waiting for goroutines to stop")
	}

	a.sendFinalMetrics()

	close(a.shutdownChan)
}

// sendFinalMetrics отправляет последние собранные метрики
func (a *Agent) sendFinalMetrics() {
	logger.Log.Info("Sending final metrics...")

	a.Mutex.Lock()
	defer a.Mutex.Unlock()

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
