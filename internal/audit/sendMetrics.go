package audit

import (
	"bytes"
	"net/http"
	"os"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

// FILE
type FileSink struct {
	path string
}

func NewFileSink(path string) *FileSink {
	return &FileSink{path: path}
}

func (s *FileSink) Notify(event Event) {
	if s.path == "" {
		return
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		logger.Log.Error("audit: OpenFile", zap.Error(err))
		return
	}
	defer f.Close()

	data, err := event.Marshal()
	if err != nil {
		return
	}

	_, err = f.Write(append(data, '\n'))
	if err != nil {
		logger.Log.Error("audit: File write", zap.Error(err))
	}
}

// HTTP
type HTTPSink struct {
	url    string
	client *http.Client
}

func NewHTTPSink(url string, timeout time.Duration) *HTTPSink {
	return &HTTPSink{
		url: url,
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

func (s *HTTPSink) Notify(event Event) {
	if s.url == "" {
		return
	}

	data, err := event.Marshal()
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, s.url, bytes.NewReader(data))
	if err != nil {
		logger.Log.Error("audit: NewRequest", zap.Error(err))
		return
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := s.client.Do(req)
	if err != nil {
		logger.Log.Error("audit: HTTP request", zap.Error(err))
	}
	resp.Body.Close()
}
