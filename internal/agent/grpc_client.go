package agent

import (
	"context"
	"fmt"
	"strings"
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
}

// NewGRPCClient создает новый gRPC клиент
func NewGRPCClient(address string, logger *zap.Logger) (*GRPCClient, error) {
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")

	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(10*time.Second),
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

	localIP := getOutboundIP()
	if localIP == "" {
		c.logger.Warn("Failed to get local IP")
		localIP = "unknown"
	}

	ctx = metadata.AppendToOutgoingContext(ctx, "x-real-ip", localIP)

	req := &pb.UpdateMetricsRequest{
		Metrics: pbMetrics,
	}

	c.logger.Info("Sending metrics via gRPC",
		zap.Int("count", len(pbMetrics)),
		zap.String("client_ip", localIP))

	_, err := c.client.UpdateMetrics(ctx, req)
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
