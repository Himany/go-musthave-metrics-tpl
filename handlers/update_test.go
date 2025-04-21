package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

func TestWebhook(t *testing.T) {
	testCases := []struct {
		name         string
		method       string
		expectedCode int
		expectedBody string
		contentType  string
		metricType   string
		metricName   string
		metricValue  string
	}{
		//Методы
		{name: "[Методы] Проверка метода GET", method: http.MethodGet, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "1"},
		{name: "[Методы] Проверка метода PUT", method: http.MethodPut, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "1"},
		{name: "[Методы] Проверка метода DELETE", method: http.MethodDelete, expectedCode: http.StatusMethodNotAllowed, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "1"},
		{name: "[Методы] Проверка метода POST", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "1"},

		//Контент
		{name: "[Контент] Тест", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "1"},
		{name: "[Контент] Json", method: http.MethodPost, expectedCode: http.StatusUnsupportedMediaType, expectedBody: "", contentType: "application/json", metricType: "gauge", metricName: "test", metricValue: "1"},
		{name: "[Контент] Пусто", method: http.MethodPost, expectedCode: http.StatusUnsupportedMediaType, expectedBody: "", contentType: "", metricType: "gauge", metricName: "test", metricValue: "1"},

		//Данные
		{name: "[Параметры] Отсутствует тип метрики", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "", metricName: "test", metricValue: "1"},
		{name: "[Параметры] Отсутствует название метрики", method: http.MethodPost, expectedCode: http.StatusNotFound, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "", metricValue: "1"},
		{name: "[Параметры] Отсутствует значение метрики", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: ""},

		//Разные типы метрик
		{name: "[Типы] Проверка типа gauge", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "12.32"},
		{name: "[Типы] Проверка типа counter", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "counter", metricName: "test", metricValue: "2"},
		{name: "[Типы] Неизвестный тип", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "blabla", metricName: "test", metricValue: "2"},

		//Значения
		{name: "[Значения] gauge строка", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "test"},
		{name: "[Значения] gauge float", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "42.2"},
		{name: "[Значения] gauge int", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: "42"},
		{name: "[Значения] counter строка", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "counter", metricName: "test", metricValue: "test"},
		{name: "[Значения] counter float", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "counter", metricName: "test", metricValue: "42.2"},
		{name: "[Значения] counter int", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "text/plain", metricType: "counter", metricName: "test", metricValue: "42"},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "/update/"+tc.metricType+"/"+tc.metricName+"/"+tc.metricValue, nil)
			r.Header.Set("Content-Type", tc.contentType)

			w := httptest.NewRecorder()

			memStorage := storage.NewMemStorage()
			handler := &Handler{Repo: memStorage}
			handler.UpdateHandler(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, fmt.Sprintf("[Код ответа не совпадает с ожидаемым] %s", tc.name))

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}
