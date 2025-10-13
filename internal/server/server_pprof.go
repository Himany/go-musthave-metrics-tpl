package server

import (
	"net/http"
	_ "net/http/pprof"
)

func init() {
	go func() { _ = http.ListenAndServe("127.0.0.1:6060", nil) }()
}
