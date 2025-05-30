package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/middleware"
)

func Run(handler *handlers.Handler, flagRunAddr string) error {
	r := chi.NewRouter()

	r.Get("/", middleware.CheckPlainTextContentType(handler.GetAllMetrics))

	r.Get("/value/{type}/{name}", middleware.CheckPlainTextContentType(handler.GetMetricQuery))
	r.Post("/value/", handler.GetMetricJSON)

	r.Post("/update/{type}/{name}/{value}", middleware.CheckPlainTextContentType(handler.UpdateHandlerQuery))
	r.Post("/update/", handler.UpdateHandlerJSON)

	return http.ListenAndServe(flagRunAddr, logger.RequestLogger(middleware.Gzip(r)))
}
