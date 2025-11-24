package server

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	pb "github.com/Himany/go-musthave-metrics-tpl/internal/proto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/repository"
)

// GRPCServer реализует gRPC сервер для метрик
type GRPCServer struct {
	pb.UnimplementedMetricsServer
	storage       repository.MetricsRepo
	logger        *zap.Logger
	trustedSubnet string
}

// NewGRPCServer создает новый gRPC сервер
func NewGRPCServer(storage repository.MetricsRepo, logger *zap.Logger, trustedSubnet string) *GRPCServer {
	return &GRPCServer{
		storage:       storage,
		logger:        logger,
		trustedSubnet: trustedSubnet,
	}
}

// UpdateMetrics обновляет метрики на сервере
func (s *GRPCServer) UpdateMetrics(ctx context.Context, req *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	s.logger.Info("Received UpdateMetrics request", zap.Int("metrics_count", len(req.Metrics)))

	metrics := make([]models.Metrics, 0, len(req.Metrics))
	for _, pbMetric := range req.Metrics {
		metric := models.Metrics{
			ID: pbMetric.Id,
		}

		switch pbMetric.Type {
		case pb.Metric_GAUGE:
			metric.MType = "gauge"
			metric.Value = &pbMetric.Value
		case pb.Metric_COUNTER:
			metric.MType = "counter"
			metric.Delta = &pbMetric.Delta
		default:
			return nil, status.Errorf(codes.InvalidArgument, "unknown metric type: %v", pbMetric.Type)
		}

		metrics = append(metrics, metric)
	}

	if err := s.storage.BatchUpdate(ctx, metrics); err != nil {
		s.logger.Error("Failed to update metrics", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "failed to update metrics: %v", err)
	}

	s.logger.Info("Successfully updated metrics", zap.Int("count", len(metrics)))
	return &pb.UpdateMetricsResponse{}, nil
}

// trustedSubnetInterceptor проверяет принадлежность IP-адреса доверенной подсети
func (s *GRPCServer) trustedSubnetInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	if s.trustedSubnet == "" {
		return handler(ctx, req)
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.PermissionDenied, "no metadata found")
	}

	realIPs := md.Get("x-real-ip")
	if len(realIPs) == 0 {
		return nil, status.Error(codes.PermissionDenied, "x-real-ip header not found")
	}

	clientIP := realIPs[0]
	s.logger.Debug("Checking client IP", zap.String("ip", clientIP), zap.String("trusted_subnet", s.trustedSubnet))

	if !s.isIPInTrustedSubnet(clientIP) {
		s.logger.Warn("Client IP not in trusted subnet", zap.String("ip", clientIP))
		return nil, status.Error(codes.PermissionDenied, "client IP not in trusted subnet")
	}

	return handler(ctx, req)
}

// isIPInTrustedSubnet проверяет, принадлежит ли IP адрес доверенной подсети
func (s *GRPCServer) isIPInTrustedSubnet(ipStr string) bool {
	clientIP, err := netip.ParseAddr(ipStr)
	if err != nil {
		s.logger.Error("Failed to parse client IP", zap.String("ip", ipStr), zap.Error(err))
		return false
	}

	_, trustedNet, err := net.ParseCIDR(s.trustedSubnet)
	if err != nil {
		s.logger.Error("Failed to parse trusted subnet", zap.String("subnet", s.trustedSubnet), zap.Error(err))
		return false
	}

	return trustedNet.Contains(net.IP(clientIP.AsSlice()))
}

// StartGRPCServer запускает gRPC сервер
func StartGRPCServer(address string, storage repository.MetricsRepo, logger *zap.Logger, trustedSubnet string) (*grpc.Server, error) {
	address = strings.TrimPrefix(address, "http://")
	address = strings.TrimPrefix(address, "https://")

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", address, err)
	}

	grpcServer := &GRPCServer{
		storage:       storage,
		logger:        logger,
		trustedSubnet: trustedSubnet,
	}

	server := grpc.NewServer(
		grpc.UnaryInterceptor(grpcServer.trustedSubnetInterceptor),
	)

	pb.RegisterMetricsServer(server, grpcServer)

	logger.Info("Starting gRPC server", zap.String("address", address))

	go func() {
		if err := server.Serve(listener); err != nil {
			logger.Error("gRPC server failed", zap.Error(err))
		}
	}()

	return server, nil
}
