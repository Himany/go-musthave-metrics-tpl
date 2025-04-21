package storage

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	repo := NewMemStorage()

	testCases := []struct {
		key           string
		value         float64
		expectedValue float64
		addValue      bool
		isOk          bool
	}{
		{key: "test1", value: 42.21, expectedValue: 42.21, addValue: true, isOk: true},
		{key: "test2", value: -42.21, expectedValue: -42.21, addValue: true, isOk: true},
		{key: "test3", value: 0, expectedValue: 0, addValue: true, isOk: true},
		{key: "test4", value: 42.21, expectedValue: 42.21, addValue: false, isOk: false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("GAUGE[%s]: %v -> %v", tc.key, tc.value, tc.expectedValue), func(t *testing.T) {
			if tc.addValue {
				repo.UpdateGauge(tc.key, tc.value)
			}

			resultValue, isOk := repo.GetGauge(tc.key)
			if isOk {
				assert.Equal(t, tc.expectedValue, resultValue, "Ошибка значения")
			}
			assert.Equal(t, tc.addValue, isOk, "Ошибка наличия")
		})
	}
}

func TestMemStorage_UpdateCounter(t *testing.T) {
	repo := NewMemStorage()

	testCases := []struct {
		key           string
		value         int64
		expectedValue int64
		addValue      bool
		isOk          bool
	}{
		{key: "test1", value: 42, expectedValue: 42, addValue: true, isOk: true},
		{key: "test2", value: -42, expectedValue: -42, addValue: true, isOk: true},
		{key: "test3", value: 0, expectedValue: 0, addValue: true, isOk: true},
		{key: "test4", value: 42, expectedValue: 42, addValue: false, isOk: false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("COUNTER[%s]: %v -> %v", tc.key, tc.value, tc.expectedValue), func(t *testing.T) {
			if tc.addValue {
				repo.UpdateCounter(tc.key, tc.value)
			}

			resultValue, isOk := repo.GetCounter(tc.key)
			if isOk {
				assert.Equal(t, tc.expectedValue, resultValue, "Ошибка значения")
			}
			assert.Equal(t, tc.addValue, isOk, "Ошибка наличия")
		})
	}
}
