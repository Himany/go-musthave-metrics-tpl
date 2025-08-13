package retry

import (
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

func WithRetry(operation func() error, isRetruableError func(error) bool, inType string) error {
	retryDelays := []int{0, 1, 3, 5}
	var lastErr error

	for attempt := 0; attempt < len(retryDelays); attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
		if !isRetruableError(err) {
			return err
		}

		logger.Log.Warn("Retriable",
			zap.String("operation", inType),
			zap.Int("attempt", attempt+1),
			zap.Error(err),
		)

		if retryDelays[attempt] != 0 {
			time.Sleep(time.Duration(retryDelays[attempt]) * time.Second)
		}
	}

	logger.Log.Error("operation failed",
		zap.String("operation", inType),
		zap.Error(lastErr),
	)

	return lastErr
}
