package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/Himany/go-musthave-metrics-tpl/internal/compress"
)

func CheckApplicationJSONContentType(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "application/json") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		h(w, r)
	}
}

func CheckPlainTextContentType(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if contentType != "" && !strings.HasPrefix(contentType, "text/plain") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			return
		}

		h(w, r)
	}
}

func Gzip(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// по умолчанию устанавливаем оригинальный http.ResponseWriter как тот,
		// который будем передавать следующей функции
		ow := w

		// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
		acceptEncoding := r.Header.Get("Accept-Encoding")
		supportsGzip := strings.Contains(acceptEncoding, "gzip")
		if supportsGzip {
			// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
			cw := compress.NewCompressWriter(w)
			// меняем оригинальный http.ResponseWriter на новый
			ow = cw
			// не забываем отправить клиенту все сжатые данные после завершения middleware
			defer cw.Close()
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		contentEncoding := r.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(contentEncoding, "gzip")
		if sendsGzip {
			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			cr, err := compress.NewCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}

func CheckHash(key string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if key == "" {
			h(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		r.Body = io.NopCloser(strings.NewReader(string(body)))

		receivedHash := r.Header.Get("HashSHA256")
		if receivedHash == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hasher := hmac.New(sha256.New, []byte(key))
		hasher.Write(body)
		expectedHash := hasher.Sum(nil)

		receivedHashBytes, err := hex.DecodeString(receivedHash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !hmac.Equal(expectedHash, receivedHashBytes) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		h(w, r)
	}
}
