package server

import (
	"net/http"

	"github.com/Himany/go-musthave-metrics-tpl/internal/crypto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/middleware"
	"github.com/Himany/go-musthave-metrics-tpl/internal/server/handlers"
	"github.com/go-chi/chi/v5"
)

func CreateRouter(handler *handlers.Handler, key string, decryptor *crypto.RSAEncryptor, trustedSubnet string) http.Handler {
	r := chi.NewRouter()

	r.Get("/", middleware.CheckPlainTextContentType(handler.GetAllMetrics))
	r.Get("/ping", handler.GetPing)
	r.Get("/value/{type}/{name}", middleware.CheckPlainTextContentType(handler.GetMetricQuery))
	r.Post("/value/", middleware.CheckApplicationJSONContentType(handler.GetMetricJSON))
	r.Post("/update/{type}/{name}/{value}", middleware.CheckPlainTextContentType(handler.UpdateHandlerQuery))

	r.With(
		middleware.CheckTrustedSubnet(trustedSubnet),
		middleware.DecryptBody(decryptor),
	).Post("/update/", middleware.CheckApplicationJSONContentType(middleware.CheckHash(key, handler.UpdateHandlerJSON)))

	r.With(
		middleware.CheckTrustedSubnet(trustedSubnet),
		middleware.DecryptBody(decryptor),
	).Post("/updates/", middleware.CheckApplicationJSONContentType(middleware.CheckHash(key, handler.BatchUpdateJSON)))

	return middleware.LoggingMiddleware(logger.RequestLogger(middleware.Gzip(r)))
}

func Router(handler *handlers.Handler, runAddr string, key string, decryptor *crypto.RSAEncryptor) error {
	router := CreateRouter(handler, key, decryptor, "")
	return http.ListenAndServe(runAddr, router)
}
