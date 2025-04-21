package storage

type Storage interface {
	UpdateGauge(name string, value float64)
	UpdateCounter(name string, value int64)
	GetGauge(name string) (float64, bool)
	GetCounter(name string) (int64, bool)
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
	val, ok := s.Gauge[name]
	return val, ok
}

func (s *MemStorage) GetCounter(name string) (int64, bool) {
	val, ok := s.Counter[name]
	return val, ok
}
