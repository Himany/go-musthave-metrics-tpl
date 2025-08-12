package agent

import (
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

func (a *agent) CreateWorkers() {
	for i := 0; i < a.RateLimit; i++ {
		go a.WorkerCreateBatchRequest()
	}
}

func (a *agent) WorkerCreateBatchRequest() {
	for batch := range a.Tasks {
		err := a.createBatchRequest(batch)
		if err != nil {
			logger.Log.Error("createBatchRequest", zap.Error(err))
		}
	}
}
