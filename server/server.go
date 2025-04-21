package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
)

func Run(handler *handlers.Handler) error {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.GetAllMetrics)
		r.Get("/value/{type}/{name}", handler.GetMetric)
		r.Post("/update/{type}/{name}/{value}", handler.UpdateHandler)
	})

	return http.ListenAndServe(":8080", r)
}
