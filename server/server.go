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
		r.Route("/update/", func(r chi.Router) {
			r.Get("/{type}/{name}", handler.GetMetric)
			r.Post("/{type}/{name}/{value}", handler.UpdateHandler)
		})
	})

	return http.ListenAndServe(":8080", r)
}
