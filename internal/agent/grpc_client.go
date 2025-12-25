package agent

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	pb "github.com/Himany/go-musthave-metrics-tpl/internal/proto"
)

// GRPCClient представляет gRPC клиент для отправки метрик
type GRPCClient struct {
	client pb.MetricsClient
	conn   *grpc.ClientConn
	logger *zap.Logger

	// Кешированный IP адрес
	cachedIP   string
	ipCacheMux sync.RWMutex
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(address string, logger *zap.Logger) (*GRPCClient, error) {
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")

	conn, err := grpc.NewClient(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %w", address, err)
	}

	client := pb.NewMetricsClient(conn)

	logger.Info("Connected to gRPC server", zap.String("address", address))

	return &GRPCClient{
		client: client,
		conn:   conn,
		logger: logger,
	}, nil
}

// getLocalIP получает IP адрес с кешированием
func (c *GRPCClient) getLocalIP() string {
	c.ipCacheMux.RLock()
	if c.cachedIP != "" {
		defer c.ipCacheMux.RUnlock()
		return c.cachedIP
	}
	c.ipCacheMux.RUnlock()

	c.ipCacheMux.Lock()
	defer c.ipCacheMux.Unlock()

	// Проверяем еще раз после получения блокировки записи
	if c.cachedIP != "" {
		return c.cachedIP
	}

	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		c.logger.Warn("Failed to get outbound IP", zap.Error(err))
		c.cachedIP = "unknown"
		return c.cachedIP
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	c.cachedIP = localAddr.IP.String()
	return c.cachedIP
}

// Close закрывает соединение с gRPC сервером
func (c *GRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// SendMetrics отправляет метрики на gRPC сервер
func (c *GRPCClient) SendMetrics(ctx context.Context, metrics []models.Metrics) error {
	if len(metrics) == 0 {
		c.logger.Debug("No metrics to send")
		return nil
	}

	pbMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		pbMetric := &pb.Metric{
			Id: metric.ID,
		}

		switch metric.MType {
		case "gauge":
			pbMetric.Type = pb.Metric_GAUGE
			if metric.Value != nil {
				pbMetric.Value = *metric.Value
			}
		case "counter":
			pbMetric.Type = pb.Metric_COUNTER
			if metric.Delta != nil {
				pbMetric.Delta = *metric.Delta
			}
		default:
			c.logger.Warn("Unknown metric type", zap.String("type", metric.MType), zap.String("id", metric.ID))
			continue
		}

		pbMetrics = append(pbMetrics, pbMetric)
	}

	if len(pbMetrics) == 0 {
		c.logger.Debug("No valid metrics to send after conversion")
		return nil
	}

	localIP := c.getLocalIP()

	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", localIP)

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	c.logger.Info("Sending metrics via gRPC",
		zap.Int("count", len(pbMetrics)),
		zap.String("client_ip", localIP))

	// Устанавливаем таймаут для gRPC запроса
	requestCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := c.client.UpdateMetrics(requestCtx, req)
	if err != nil {
		c.logger.Error("Failed to send metrics via gRPC", zap.Error(err))
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	c.logger.Info("Successfully sent metrics via gRPC", zap.Int("count", len(pbMetrics)))
	return nil
}

// SendMetricsBatch отправляет метрики батчами
func (c *GRPCClient) SendMetricsBatch(ctx context.Context, metrics []models.Metrics, batchSize int) error {
	if batchSize <= 0 {
		batchSize = 100
	}

	c.logger.Info("Sending metrics in batches",
		zap.Int("total_metrics", len(metrics)),
		zap.Int("batch_size", batchSize))

	for i := 0; i < len(metrics); i += batchSize {
		end := i + batchSize
		if end > len(metrics) {
			end = len(metrics)
		}

		batch := metrics[i:end]

		c.logger.Debug("Sending batch",
			zap.Int("batch_number", i/batchSize+1),
			zap.Int("batch_size", len(batch)))

		if err := c.SendMetrics(ctx, batch); err != nil {
			return fmt.Errorf("failed to send batch %d: %w", i/batchSize+1, err)
		}

		if end < len(metrics) {
			time.Sleep(10 * time.Millisecond)
		}
	}

	c.logger.Info("All batches sent successfully")
	return nil
}
