package main

import (
	"github.com/Himany/go-musthave-metrics-tpl/handlers"
	"github.com/Himany/go-musthave-metrics-tpl/server"
	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

func main() {
	parseFlags()

	memStorage := storage.NewMemStorage()
	handler := &handlers.Handler{Repo: memStorage}

	if err := server.Run(handler, runAddr); err != nil {
		panic(err)
	}
}
