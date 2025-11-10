package audit

import (
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

// Event — формат события аудита
type Event struct {
	TS        int64    `json:"ts"`
	Metrics   []string `json:"metrics"`
	IPAddress string   `json:"ip_address"`
}

type Observer interface {
	Notify(event Event)
}

type Publisher struct {
	observers []Observer
}

func NewPublisher() *Publisher {
	return &Publisher{observers: make([]Observer, 0, 2)}
}

func (p *Publisher) Register(obs Observer) {
	if obs == nil {
		return
	}
	p.observers = append(p.observers, obs)
}

func (p *Publisher) Publish(event Event) {
	for _, obs := range p.observers {
		o := obs
		e := event
		go o.Notify(e)
	}
}

func BuildEvent(r *http.Request, metricNames []string) Event {
	return Event{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: clientIP(r),
	}
}

func clientIP(r *http.Request) string {
	// приоритет X-Real-IP
	if ip := strings.TrimSpace(r.Header.Get("X-Real-IP")); ip != "" {
		return ip
	}
	// затем первый из X-Forwarded-For
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			ip := strings.TrimSpace(parts[0])
			if ip != "" {
				return ip
			}
		}
	}
	// иначе из RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil && host != "" {
		return host
	}
	return r.RemoteAddr
}

func (e Event) Marshal() ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		logger.Log.Error("audit: Marshal", zap.Error(err))
		return nil, err
	}
	return b, nil
}
