package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
)

func Run(handler *handlers.Handler, flagRunAddr string) error {
	r := chi.NewRouter()

	r.Get("/", handler.GetAllMetrics)
	r.Get("/value/{type}/{name}", handler.GetMetric)
	r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)

	return http.ListenAndServe(flagRunAddr, logger.RequestLogger(r))
}
