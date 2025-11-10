package models

import "strconv"

// Metrics описывает структуру метрики, передаваемую между агентом и сервером.
// generate:reset
type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

// FormatGaugeValue форматирует значение gauge метрики в строку
func FormatGaugeValue(value float64) string {
	return strconv.FormatFloat(value, 'g', -1, 64)
}

// FormatCounterValue форматирует значение counter метрики в строку
func FormatCounterValue(value int64) string {
	return strconv.FormatInt(value, 10)
}
