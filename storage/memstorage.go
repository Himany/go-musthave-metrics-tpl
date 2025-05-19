package storage

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetKeyGauge() []string
	GetKeyCounter() []string
	SaveData() error
	LoadData() error
	SaveHandler(int)
}

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64

	fileToSave string
	isSyncSave bool
}

func NewMemStorage(path string, isSyncSave bool) *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),

		fileToSave: path,
		isSyncSave: isSyncSave,
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.Gauge[name] = value
	if err := s.SaveData(); err != nil {
		logger.Log.Error("UpdateGauge", zap.Error(err))
	}
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.Counter[name] = value
	if err := s.SaveData(); err != nil {
		logger.Log.Error("UpdateGauge", zap.Error(err))
	}
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	val, ok := s.Gauge[name]
	return val, ok
}

func (s *MemStorage) GetKeyGauge() []string {
	keys := make([]string, 0, len(s.Gauge))
	for key := range s.Gauge {
		keys = append(keys, key)
	}
	return keys
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := s.Counter[name]
	return val, ok
}

func (s *MemStorage) GetKeyCounter() []string {
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

func (s *MemStorage) SaveData() error {
	if s.fileToSave == "" {
		return errors.New("file is not specified")
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

	return nil
}

func (s *MemStorage) LoadData() error {
	if s.fileToSave == "" {
		return errors.New("file is not specified")
	}

	var save saveFormat

	data, err := os.ReadFile(s.fileToSave)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(data, &save); err != nil {
		return err
	}

	s.Gauge = save.Gauge
	s.Counter = save.Counter

	return nil
}

func (s *MemStorage) SaveHandler(interval int) {
	for {
		if err := s.SaveData(); err != nil {
			logger.Log.Error("SaveHandler", zap.Error(err))
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}
}
