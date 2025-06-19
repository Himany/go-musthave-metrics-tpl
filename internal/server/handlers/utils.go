package handlers

import "strconv"

func (h *Handler) getStringValue(metricType string, metricName string) (string, bool) {
	result := ""

	switch metricType {
	case "gauge":
		value, ok := h.Repo.GetGauge(metricName)
		if !ok {
			return "", false
		}
		result = strconv.FormatFloat(value, 'f', -1, 64)
		return result, ok

	case "counter":
		value, ok := h.Repo.GetCounter(metricName)
		if ok {
			result = strconv.FormatInt(value, 10)
		}
		return result, ok

	default:
		return "", false
	}
}
