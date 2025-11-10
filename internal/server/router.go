package server

import (
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/middleware"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/go-chi/chi/v5"
)

func Router(handler *handlers.Handler, runAddr string, key string) error {
	r := chi.NewRouter()

	r.Get("/", middleware.CheckPlainTextContentType(handler.GetAllMetrics))

	r.Get("/ping", handler.GetPing)

	r.Get("/value/{type}/{name}", middleware.CheckPlainTextContentType(handler.GetMetricQuery))
	r.Post("/value/", middleware.CheckApplicationJSONContentType(handler.GetMetricJSON))

	r.Post("/update/{type}/{name}/{value}", middleware.CheckPlainTextContentType(handler.UpdateHandlerQuery))
	r.Post("/update/", middleware.CheckApplicationJSONContentType(middleware.CheckHash(key, handler.UpdateHandlerJSON)))

	r.Post("/updates/", middleware.CheckApplicationJSONContentType(middleware.CheckHash(key, handler.BatchUpdateJSON)))

	return http.ListenAndServe(runAddr, middleware.LoggingMiddleware(logger.RequestLogger(middleware.Gzip(r))))
}
