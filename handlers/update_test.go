package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"

	"github.com/Himany/go-musthave-metrics-tpl/storage"
)

func TestUpdate(t *testing.T) {
	memStorage := storage.NewMemStorage("", false)
	handler := &Handler{Repo: memStorage}

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateHandlerQuery)

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
		{name: "[Контент] Пусто", method: http.MethodPost, expectedCode: http.StatusOK, expectedBody: "", contentType: "", metricType: "gauge", metricName: "test", metricValue: "1"},

		//Данные
		{name: "[Параметры] Отсутствует тип метрики", method: http.MethodPost, expectedCode: http.StatusBadRequest, expectedBody: "", contentType: "text/plain", metricType: "", metricName: "test", metricValue: "1"},
		{name: "[Параметры] Отсутствует название метрики", method: http.MethodPost, expectedCode: http.StatusNotFound, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "", metricValue: "1"},
		{name: "[Параметры] Отсутствует значение метрики", method: http.MethodPost, expectedCode: http.StatusNotFound, expectedBody: "", contentType: "text/plain", metricType: "gauge", metricName: "test", metricValue: ""},

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
			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedCode, w.Code, fmt.Sprintf("[Код ответа не совпадает с ожидаемым] %s", tc.name))

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, w.Body.String(), "Тело ответа не совпадает с ожидаемым")
			}
		})
	}
}

func TestGetMetric(t *testing.T) {
	memStorage := storage.NewMemStorage("", false)
	handler := &Handler{Repo: memStorage}

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateHandlerQuery)
	router.Get("/value/{type}/{name}", handler.GetMetricQuery)

	testCases := []struct {
		name            string
		expectedCodeAdd int
		expectedCodeGet int
		metricType      string
		metricName      string
		metricValue     string
	}{
		{name: "Добавления ноля (float64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "gauge", metricName: "f1", metricValue: "0"},
		{name: "Добавления положительного целого числа (float64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "gauge", metricName: "f2", metricValue: "42"},
		{name: "Добавления отрицательного целого числа (float64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "gauge", metricName: "f3", metricValue: "-42"},
		{name: "Добавления положительного числа с плавающей точкой (float64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "gauge", metricName: "f4", metricValue: "42.2"},
		{name: "Добавления отрицательного числа с плавающей точкой (float64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "gauge", metricName: "f5", metricValue: "-42.2"},
		{name: "Добавления ноля (int64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "counter", metricName: "i1", metricValue: "0"},
		{name: "Добавления положительного целого числа (int64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "counter", metricName: "i2", metricValue: "42"},
		{name: "Добавления отрицательного целого числа (int64)", expectedCodeAdd: http.StatusOK, expectedCodeGet: http.StatusOK, metricType: "counter", metricName: "i3", metricValue: "-42"},
		{name: "Добавления положительного числа с плавающей точкой (int64)", expectedCodeAdd: http.StatusBadRequest, expectedCodeGet: http.StatusNotFound, metricType: "counter", metricName: "i4", metricValue: "42.2"},
		{name: "Добавления отрицательного числа с плавающей точкой (int64)", expectedCodeAdd: http.StatusBadRequest, expectedCodeGet: http.StatusNotFound, metricType: "counter", metricName: "i5", metricValue: "-42.2"},
	}

	for index, tc := range testCases {
		t.Run(fmt.Sprintf("TestGetMetric: #%d", index), func(t *testing.T) {
			//post
			rPost := httptest.NewRequest(http.MethodPost, "/update/"+tc.metricType+"/"+tc.metricName+"/"+tc.metricValue, nil)
			rPost.Header.Set("Content-Type", "text/plain")

			wPost := httptest.NewRecorder()
			router.ServeHTTP(wPost, rPost)

			assert.Equal(t, tc.expectedCodeAdd, wPost.Code, fmt.Sprintf("[Код ответа не совпадает с ожидаемым (ADD)] %s", tc.name))

			//get
			rGet := httptest.NewRequest(http.MethodGet, "/value/"+tc.metricType+"/"+tc.metricName, nil)

			wGet := httptest.NewRecorder()
			router.ServeHTTP(wGet, rGet)

			assert.Equal(t, tc.expectedCodeGet, wGet.Code, fmt.Sprintf("[Код ответа не совпадает с ожидаемым (GET)] %s", tc.name))
		})
	}
}

func TestGetAllMetrics(t *testing.T) {
	memStorage := storage.NewMemStorage("", false)
	handler := &Handler{Repo: memStorage}

	router := chi.NewRouter()
	router.Post("/update/{type}/{name}/{value}", handler.UpdateHandlerQuery)
	router.Get("/", handler.GetAllMetrics)

	// Метрики для добавления
	metrics := []struct {
		metricType  string
		metricName  string
		metricValue string
	}{
		{metricType: "gauge", metricName: "TestGaugeMetric", metricValue: "100.5"},
		{metricType: "counter", metricName: "TestCounterMetric", metricValue: "42"},
	}

	for _, m := range metrics {
		req := httptest.NewRequest(http.MethodPost, "/update/"+m.metricType+"/"+m.metricName+"/"+m.metricValue, nil)
		req.Header.Set("Content-Type", "text/plain")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "[POST] Код ответа не совпадает при добавлении метрики %s", m.metricName)
	}

	t.Run("[Метрики] Проверка получения всех метрик", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "[GET /] Код ответа не совпадает")

		body := w.Body.String()

		assert.Contains(t, body, "TestGaugeMetric: 100.5;", "Отсутствует gauge-метрика в списке")
		assert.Contains(t, body, "TestCounterMetric: 42;", "Отсутствует counter-метрика в списке")
	})
}
