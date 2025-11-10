package errors

import (
	"errors"
)

// Ошибки валидации данных
var (
	ErrFieldIDRequired    = errors.New("field 'id' is required")
	ErrFieldTypeInvalid   = errors.New("field 'type' must have a value of 'gauge' or 'counter'")
	ErrFieldValueRequired = errors.New("field 'value' is required (with the 'gauge' type)")
	ErrFieldDeltaRequired = errors.New("field 'delta' is required (with the 'counter' type)")
)

// Ошибки бизнес-логики сервиса метрик
var (
	ErrMetricNotFound       = errors.New("metric not found")
	ErrUnknownMetricType    = errors.New("unknown metric type")
	ErrGaugeValueRequired   = errors.New("gauge value is required")
	ErrCounterDeltaRequired = errors.New("counter delta is required")
	ErrEmptyMetrics         = errors.New("empty metrics")
	ErrMetricIDRequired     = errors.New("metric ID is required")
	ErrInvalidMetricType    = errors.New("invalid metric type")
)
