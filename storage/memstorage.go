package storage

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
	"go.uber.org/zap"
)

type MemStorage interface {
	Ping() error
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetKeyGauge() []string
	GetKeyCounter() []string
	SaveData() error
	LoadData() error
	SaveHandler(int)
	BatchUpdate(metrics []models.Metrics) error
}

type MemStorageData struct {
	Gauge   map[string]float64
	Counter map[string]int64

	fileToSave string
	isSyncSave bool
}

func NewMemStorage(path string, isSyncSave bool) *MemStorageData {
	return &MemStorageData{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),

		fileToSave: path,
		isSyncSave: isSyncSave,
	}
}

func (s *MemStorageData) Ping() error {
	return nil
}

func (s *MemStorageData) UpdateGauge(name string, value float64) {
	s.Gauge[name] = value
	if s.isSyncSave {
		if err := s.SaveData(); err != nil {
			logger.Log.Error("MEM UpdateGauge", zap.Error(err))
		}
	}
}

func (s *MemStorageData) UpdateCounter(name string, value int64) {
	s.Counter[name] = value
	if s.isSyncSave {
		if err := s.SaveData(); err != nil {
			logger.Log.Error("MEM UpdateGauge", zap.Error(err))
		}
	}
}

func (s *MemStorageData) GetGauge(name string) (float64, bool) {
	val, ok := s.Gauge[name]
	return val, ok
}

func (s *MemStorageData) GetKeyGauge() []string {
	keys := make([]string, 0, len(s.Gauge))
	for key := range s.Gauge {
		keys = append(keys, key)
	}
	return keys
}

func (s *MemStorageData) GetCounter(name string) (int64, bool) {
	val, ok := s.Counter[name]
	return val, ok
}

func (s *MemStorageData) GetKeyCounter() []string {
	keys := make([]string, 0, len(s.Counter))
	for key := range s.Counter {
		keys = append(keys, key)
	}
	return keys
}

// files
type saveFormat struct {
	Gauge   map[string]float64 `json:"gauge"`
	Counter map[string]int64   `json:"counter"`
}

func (s *MemStorageData) SaveData() error {
	if s.fileToSave == "" {
		return errors.New("MEM file is not specified")
	}

	// создаем файл
	file, err := os.OpenFile(s.fileToSave, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// сериализуем структуру в JSON формат
	data, err := json.Marshal(saveFormat{
		Gauge:   s.Gauge,
		Counter: s.Counter,
	})
	if err != nil {
		return err
	}

	// сохраняем данные в файл
	_, err = file.Write(data)
	if err != nil {
		return err
	}

	logger.Log.Info("MEM metrics saved successfully", zap.String("path", s.fileToSave))

	return nil
}

func (s *MemStorageData) LoadData() error {
	if s.fileToSave == "" {
		return errors.New("MEM file is not specified")
	}

	var save saveFormat

	data, err := os.ReadFile(s.fileToSave)
	if os.IsNotExist(err) {
		logger.Log.Warn("MEM metrics file not found, skipping restore", zap.String("path", s.fileToSave))
		return nil
	}
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &save); err != nil {
		return err
	}

	s.Gauge = save.Gauge
	s.Counter = save.Counter

	logger.Log.Info("MEM metrics loaded successfully", zap.String("path", s.fileToSave))
	return nil
}

func (s *MemStorageData) SaveHandler(interval int) {
	for {
		if err := s.SaveData(); err != nil {
			logger.Log.Error("MEM SaveHandler", zap.Error(err))
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}

func (s *MemStorageData) BatchUpdate(metrics []models.Metrics) error {
	for _, m := range metrics {
		switch m.MType {
		case "gauge":
			if m.Value == nil {
				continue
			}
			s.Gauge[m.ID] = *m.Value

		case "counter":
			if m.Delta == nil {
				continue
			}
			s.Counter[m.ID] = *m.Delta
		default:
			logger.Log.Warn("BatchUpdate unknown metric type", zap.String("type", m.MType))
		}
	}

	return nil
}
