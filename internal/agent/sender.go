package agent

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func (a *agent) retryGzipJSONRequest(body []byte, route string, hash []byte) (*resty.Response, time.Duration, error) {
	retryDelays := []int{0, 1, 3, 5}
	var lastErr error
	var lastResp *resty.Response

	start := time.Now()
	for attempt := 0; attempt < len(retryDelays); attempt++ {
		request := a.Client.R().
			SetHeader("Content-Encoding", "gzip").
			SetHeader("Content-Type", "application/json").
			SetBody(body)

		if hash != nil {
			request.SetHeader("HashSHA256", hex.EncodeToString(hash))
		}

		resp, err := request.Post(a.URL + route)

		if err == nil {
			duration := time.Since(start)

			return resp, duration, nil
		} else {
			lastErr = err
			lastResp = resp
		}

		nonBreakingCodes := map[int]bool{
			502: true,
			503: true,
			504: true,
			429: true,
		}
		if _, nonBreaking := nonBreakingCodes[resp.StatusCode()]; resp != nil && !nonBreaking {
			break
		}

		if retryDelays[attempt] != 0 {
			time.Sleep(time.Duration(retryDelays[attempt]) * time.Second)
		}
	}

	return lastResp, time.Since(start), lastErr
}

func (a *agent) createBatchRequest(metrics []models.Metrics) error {
	if len(metrics) == 0 {
		return errors.New("empty metrics")
	}

	jsonData, err := json.Marshal(metrics)
	if err != nil {
		return err
	}

	hash := bodySignature(jsonData, a.Key)

	body, err := compressBody(jsonData)
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

func (a *agent) createRequest(metricType string, name string, delta *int64, value *float64) error {
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

	hash := bodySignature(jsonData, a.Key)

	body, err := compressBody(jsonData)
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

func (a *agent) reportHandler() {
	for {
		a.Mutex.Lock()

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

		a.Mutex.Unlock()
		err := a.createBatchRequest(batch)
		if err != nil {
			logger.Log.Error("createBatchRequest", zap.Error(err))
		}

		time.Sleep(time.Duration(a.ReportInterval) * time.Second)
	}
}
