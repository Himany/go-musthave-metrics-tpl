package server

import (
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
)

func Run(handler *handlers.Handler) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/update/", handler.UpdateHandler)
	return http.ListenAndServe(":8080", mux)
}
