package audit

import (
	"os"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/go-resty/resty/v2"
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
	client *resty.Client
}

func NewHTTPSink(url string, timeout time.Duration) *HTTPSink {
	return &HTTPSink{
		url:    url,
		client: resty.New().SetTimeout(timeout),
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

	resp, err := s.client.R().
		SetHeader("Content-Type", "application/json; charset=utf-8").
		SetBody(data).
		Post(s.url)
	if err != nil {
		logger.Log.Error("audit: HTTP request", zap.Error(err))
		return
	}
	_ = resp
}
