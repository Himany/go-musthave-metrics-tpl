package handlers

import (
	"errors"

	"github.com/Himany/go-musthave-metrics-tpl/internal/models"
)

func validateUpdateJSON(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !(metrics.MType == "gauge" || metrics.MType == "counter") {
		return errors.New("field 'type' must have a value of 'gauge' or 'counter'")
	}

	switch metrics.MType {
	case "gauge":
		if metrics.Value == nil {
			return errors.New("field 'value' is required (with the 'gauge' type)")
		}
	case "counter":
		if metrics.Delta == nil {
			return errors.New("field 'delta' is required (with the 'counter' type)")
		}
	}

	return nil
}

func validateGetMetricJSON(metrics models.Metrics) error {
	if metrics.ID == "" {
		return errors.New("field 'id' is required")
	}

	if !(metrics.MType == "gauge" || metrics.MType == "counter") {
		return errors.New("field 'type' must have a value of 'gauge' or 'counter'")
	}

	return nil
}
