package main

import (
	"flag"
	"log"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	Address string `env:"ADDRESS"`
}

var runAddr string

func parseFlags() {
	var flagRunAddr = flag.String("a", "localhost:8080", "address and port to run server")
	flag.Parse()

	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	runAddr = *flagRunAddr
	if cfg.Address != "" {
		runAddr = cfg.Address
	}
}
