package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Himany/go-musthave-metrics-tpl/internal/compress"
	"github.com/Himany/go-musthave-metrics-tpl/internal/crypto"
	"github.com/Himany/go-musthave-metrics-tpl/internal/logger"
	"go.uber.org/zap"
)

// CheckApplicationJSONContentType проверяет Content-Type и разрешает только application/json.
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

// CheckPlainTextContentType проверяет Content-Type и разрешает только text/plain.
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
		r.Body = io.NopCloser(bytes.NewReader(body))

		receivedHash := r.Header.Get("HashSHA256")
		if receivedHash == "" || receivedHash == "none" {
			h(w, r)
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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggingResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// LoggingMiddleware логирует HTTP-запросы и ответы с захватом status code
func LoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		lw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		h.ServeHTTP(lw, r)

		duration := time.Since(start)

		// Логируем запрос
		logger.Log.Info("HTTP request",
			zap.String("method", r.Method),
			zap.String("uri", r.RequestURI),
			zap.Int("status", lw.statusCode),
			zap.Duration("duration", duration),
		)
	})
}

// DecryptBody дешифрует тело запроса если включено шифрование
func DecryptBody(decryptor *crypto.RSAEncryptor) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if decryptor == nil || !decryptor.IsEnabled() {
				h.ServeHTTP(w, r)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			r.Body.Close()

			// Дешифруем данные
			decryptedBody, err := decryptor.Decrypt(body)
			if err != nil {
				logger.Log.Debug("Failed to decrypt body", zap.Error(err))
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Заменяем тело запроса дешифрованными данными
			r.Body = io.NopCloser(bytes.NewReader(decryptedBody))

			h.ServeHTTP(w, r)
		})
	}
}
