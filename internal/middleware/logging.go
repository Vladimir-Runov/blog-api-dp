package middleware

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Ключ для хранения RequestID в контексте
type ctxKey string

const RequestIDKey ctxKey = "request_id"

// LoggingMiddleware предоставляет middleware для логирования с RequestID
type LoggingMiddleware struct {
	logger *log.Logger
}

// NewLoggingMiddleware создает новый экземпляр middleware
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{logger: logger}
}

// Logger логирует каждый запрос
func (m *LoggingMiddleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter, чтобы отследить статус
		ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(ww, r)

		// Безопасное извлечение request_id
		requestID := "unknown"
		if raw := r.Context().Value(RequestIDKey); raw != nil {
			if id, ok := raw.(string); ok {
				requestID = id
			}
		}

		m.logger.Printf(
			"%s | %s | %s %s | %d | %v",
			requestID,
			r.RemoteAddr,
			r.Method,
			r.URL.Path,
			ww.statusCode,
			time.Since(start),
		)
	})
}

// RequestID генерирует и добавляет уникальный ID к каждому запросу
func (m *LoggingMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Вспомогательные структуры
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// generateRequestID генерирует простой ID (в проде, наверное, лучше будет использовать UUID)
func generateRequestID() string {
	return time.Now().UTC().Format("20060102150405.000") + "-" + strconv.Itoa(int(time.Now().UnixNano()%1000))
}
