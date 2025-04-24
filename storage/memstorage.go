package storage

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
	GetKeyGauge() []string
	GetKeyCounter() []string
}

type MemStorage struct {
	Gauge   map[string]float64
	Counter map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		Gauge:   make(map[string]float64),
		Counter: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(name string, value float64) {
	s.Gauge[name] = value
}

func (s *MemStorage) UpdateCounter(name string, value int64) {
	s.Counter[name] = value
}

func (s *MemStorage) GetGauge(name string) (float64, bool) {
	return s.Gauge[name]
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
