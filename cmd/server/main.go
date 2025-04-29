package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/server"
	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

func main() {
	parseFlags()

	memStorage := storage.NewMemStorage()
	handler := &handlers.Handler{Repo: memStorage}

	if err := server.Run(handler, runAddr); err != nil {
		log.Fatal(err)
	}
}
