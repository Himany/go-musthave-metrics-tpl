package server

import (
	"errors"
	"net/http"
	_ "net/http/pprof"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

func startPprof(addr string) {
	if addr == "" {
		return
	}

	go func() {
		if err := http.ListenAndServe(addr, nil); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Error("pprof server failed", zap.Error(err))
		}
	}()
}
