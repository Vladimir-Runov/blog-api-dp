package middleware

// github.com/google/uuid v1.6.0
import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

func TestNewLoggingMiddleware(t *testing.T) {
	logger := log.New(&bytes.Buffer{}, "", 0)

	middleware := NewLoggingMiddleware(logger)

	if middleware == nil {
		t.Fatal("NewLoggingMiddleware() вернул nil")
	}

	if middleware.logger != logger {
		t.Fatal("middleware.logger не совпадает с переданным logger")
	}
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
	var receivedRequestID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, ok := r.Context().Value(RequestIDKey).(string)
		if !ok {
			t.Fatalf("RequestID отсутствует в контексте или имеет неправильный тип")
		}

		receivedRequestID = id
		w.WriteHeader(http.StatusNoContent)
	})

	handler := NewLoggingMiddleware(log.Default()).RequestID(next)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if receivedRequestID == "" {
		t.Fatal("RequestID пустой")
	}

	if !regexp.MustCompile(`^\d{14}\.\d{3}-\d+$`).MatchString(receivedRequestID) {
		t.Errorf("неожиданный формат RequestID: %q", receivedRequestID)
	}

	if got := rec.Header().Get("X-Request-ID"); got != receivedRequestID {
		t.Errorf("X-Request-ID = %q, ожидалось %q", got, receivedRequestID)
	}

	if rec.Code != http.StatusNoContent {
		t.Errorf("статус = %d, ожидалось %d", rec.Code, http.StatusNoContent)
	}
}

func TestLoggingMiddleware_Logger(t *testing.T) {
	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	handler := NewLoggingMiddleware(logger).Logger(next)

	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.RemoteAddr = "192.0.2.1:12345"
	req = req.WithContext(
		context.WithValue(req.Context(), RequestIDKey, "request-123"),
	)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("статус = %d, ожидалось %d", rec.Code, http.StatusCreated)
	}

	logLine := logs.String()

	for _, expected := range []string{
		"request-123",
		"192.0.2.1:12345",
		"POST /users",
		"201",
	} {
		if !strings.Contains(logLine, expected) {
			t.Errorf("лог не содержит %q: %q", expected, logLine)
		}
	}
}

func TestLoggingMiddleware_LoggerUsesUnknownRequestID(t *testing.T) {
	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	handler := NewLoggingMiddleware(logger).Logger(next)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.RemoteAddr = "127.0.0.1:8080"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("статус = %d, ожидалось %d", rec.Code, http.StatusOK)
	}

	logLine := logs.String()

	for _, expected := range []string{
		"unknown",
		"127.0.0.1:8080",
		"GET /health",
		"200",
	} {
		if !strings.Contains(logLine, expected) {
			t.Errorf("лог не содержит %q: %q", expected, logLine)
		}
	}
}

// TestLoggingMiddleware_FullChain проверяет совместную работу двух middleware:
//  1. RequestID — создаёт идентификатор запроса.
//  2. Logger — логирует информацию о запросе и ответе.
//     создаётся идентификатор, передаётся дальше по цепочке, попадает в ответ и записывается в лог.
func TestLoggingMiddleware_FullChain(t *testing.T) {

	var logs bytes.Buffer
	logger := log.New(&logs, "", 0)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Context().Value(RequestIDKey) == nil {
			t.Error("RequestID отсутствует в контексте")
		}

		w.WriteHeader(http.StatusAccepted)
	})

	handler := NewLoggingMiddleware(logger).RequestID(
		NewLoggingMiddleware(logger).Logger(next),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/items", nil)
	req.RemoteAddr = "203.0.113.10:4567"

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	requestID := rec.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("заголовок X-Request-ID отсутствует")
	}

	if rec.Code != http.StatusAccepted {
		t.Errorf("статус = %d, ожидалось %d", rec.Code, http.StatusAccepted)
	}

	logLine := logs.String()

	for _, expected := range []string{
		requestID,
		"203.0.113.10:4567",
		"GET /api/items",
		"202",
	} {
		if !strings.Contains(logLine, expected) {
			t.Errorf("лог не содержит %q: %q", expected, logLine)
		}
	}
}
