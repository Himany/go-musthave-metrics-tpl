package agent

import (
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

func (a *Agent) CreateWorkers() {
	for i := 0; i < a.RateLimit; i++ {
		a.wg.Add(1)
		go a.WorkerCreateBatchRequest()
	}
}

func (a *Agent) WorkerCreateBatchRequest() {
	defer a.wg.Done()

	for {
		select {
		case <-a.ctx.Done():
			a.processRemainingTasks()
			return
		case batch := <-a.Tasks:
			err := a.createBatchRequest(batch)
			if err != nil {
				logger.Log.Error("createBatchRequest", zap.Error(err))
			}
		}
	}
}

// processRemainingTasks обрабатывает оставшиеся задачи в канале при завершении
func (a *Agent) processRemainingTasks() {
	for {
		select {
		case batch := <-a.Tasks:
			err := a.createBatchRequest(batch)
			if err != nil {
				logger.Log.Error("createBatchRequest during shutdown", zap.Error(err))
			}
		default:
			return
		}
	}
}
