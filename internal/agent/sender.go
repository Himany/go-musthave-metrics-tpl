package agent

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/Himany/go-musthave-metrics-tpl/internal/retry"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

var ErrEmptyMetrics = errors.New("empty metrics")

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func (a *Agent) retryGzipJSONRequest(body []byte, route string, hash []byte) (*resty.Response, time.Duration, error) {
	start := time.Now()
	var lastResp *resty.Response

	err := retry.WithRetry(func() error {
		request := a.Client.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetHeader("X-Real-IP", getOutboundIP()).
			SetBody(body)

		if hash != nil {
			request.SetHeader("HashSHA256", hex.EncodeToString(hash))
		}

		resp, reqErr := request.Post(a.URL + route)
		lastResp = resp
		return reqErr
	}, func(err error) bool {
		if err == nil {
			return false
		}
		if lastResp != nil {
			switch lastResp.StatusCode() {
			case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusTooManyRequests:
				return true
			default:
				return false
			}
		}
		var restyErr *resty.ResponseError
		if errors.As(err, &restyErr) {
			return true
		}
		return true
	}, "agent_http_request")

	duration := time.Since(start)
	return lastResp, duration, err
}

func (a *Agent) createBatchRequest(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return ErrEmptyMetrics
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	encryptedData, err := a.Encryptor.Encrypt(jsonData)
	if err != nil {
		return err
	}

	hash := bodySignature(encryptedData, a.Key)

	body, err := compressBody(encryptedData)
	if err != nil {
		return err
	}

	var route = "/updates/"
	resp, duration, err := a.retryGzipJSONRequest(body, route, hash)

	if err == nil && resp != nil {
		logger.Log.Info("HTTP BATCH",
			zap.String("uri (request)", a.URL+route),
			zap.String("method (request)", "POST"),
			zap.Duration("duration", duration),
			zap.Int("status (answer)", resp.StatusCode()),
			zap.Int("size (answer)", len(resp.Body())),
			zap.String("body (answer)", resp.String()),
		)
		return nil
	}

	return err
}

func (a *Agent) createRequest(metricType string, name string, delta *int64, value *float64) error {
	metrics := models.Metrics{
		ID:    name,
		MType: metricType,
		Delta: delta,
		Value: value,
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	encryptedData, err := a.Encryptor.Encrypt(jsonData)
	if err != nil {
		return err
	}

	hash := bodySignature(encryptedData, a.Key)

	body, err := compressBody(encryptedData)
	if err != nil {
		return err
	}

	var route = "/update/"
	resp, duration, err := a.retryGzipJSONRequest(body, route, hash)

	if err == nil && resp != nil {
		logger.Log.Info("HTTP request",
			zap.String("uri", a.URL+route),
			zap.String("method", "POST"),
			zap.Duration("duration", duration),
		)

		logger.Log.Info("HTTP answer",
			zap.Int("status", resp.StatusCode()),
			zap.Int("size", len(resp.Body())),
			zap.String("body", resp.String()),
		)

		return nil
	}

	return err
}

func (a *Agent) reportHandler() {
	defer a.wg.Done()

	ticker := time.NewTicker(time.Duration(a.ReportInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.mutex.Lock()

			var batch []models.Metrics
			for key, value := range a.Metrics {
				val := value
				batch = append(batch, models.Metrics{
					ID:    key,
					MType: "gauge",
					Value: &val,
				})
			}
			batch = append(batch, models.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: &a.PollCount,
			})

			a.mutex.Unlock()

			select {
			case <-a.ctx.Done():
				return
			case a.Tasks <- batch:
			}
		}
	}
}
