package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"strconv"
)

func (h *Handler) getStringValue(metricType string, metricName string) (string, bool) {
	result := ""

	switch metricType {
	case "gauge":
		value, ok := h.Storage.Repo.GetGauge(metricName)
		if !ok {
			return "", false
		}
		result = strconv.FormatFloat(value, 'f', -1, 64)
		return result, ok

	case "counter":
		value, ok := h.Storage.Repo.GetCounter(metricName)
		if ok {
			result = strconv.FormatInt(value, 10)
		}
		return result, ok

	default:
		return "", false
	}
}

func bodySignature(jsonData []byte, key string) []byte {
	if key == "" {
		return nil
	}

	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write(jsonData)
	return hasher.Sum(nil)
}
