package middleware

// go test ./internal/middleware
// go test ./internal/middleware -cover

//Тесты проверяют:
//	• захват статус-кода в responseWriter;
//	• логирование метода, URL, IP и статуса;
//	• статус 200 OK, если обработчик явно не вызвал WriteHeader;
//	• восстановление после паники;
//	• запись stack trace в лог;
//	• CORS-заголовки;
//	• обработку preflight-запроса OPTIONS;
//	• генерацию UUID в X-Request-ID;
//	• передачу request ID через контекст;
//	• сохранение исходного контекста запроса.

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func testLogger() (*log.Logger, *bytes.Buffer) {
	var buffer bytes.Buffer

	logger := log.New(
		&buffer,
		"",
		0,
	)

	return logger, &buffer
}

func TestResponseWriter_WriteHeader(t *testing.T) {
	recorder := httptest.NewRecorder()

	writer := &responseWriter{
		ResponseWriter: recorder,
		statusCode:     http.StatusOK,
	}

	writer.WriteHeader(http.StatusCreated)
	writer.WriteHeader(http.StatusBadRequest)

	if writer.statusCode != http.StatusCreated {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusCreated,
			writer.statusCode,
		)
	}

	if !writer.written {
		t.Fatal("expected writer.written to be true")
	}

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected recorder status %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}
}

func TestLoggingMiddleware_Logger(t *testing.T) {
	logger, logs := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handlerCalled := false

	handler := middleware.Logger(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		handlerCalled = true
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	request := httptest.NewRequest(
		http.MethodPost,
		"/users",
		nil,
	)
	request.RemoteAddr = "127.0.0.1:12345"

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !handlerCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusCreated {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusCreated,
			recorder.Code,
		)
	}

	logOutput := logs.String()

	for _, expected := range []string{
		"POST",
		"/users",
		"127.0.0.1:12345",
		"201",
	} {
		if !strings.Contains(logOutput, expected) {
			t.Errorf(
				"expected log to contain %q, got %q",
				expected,
				logOutput,
			)
		}
	}
}

func TestLoggingMiddleware_Logger_DefaultStatus(t *testing.T) {
	logger, logs := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handler := middleware.Logger(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		_, _ = w.Write([]byte("ok"))
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}

	if !strings.Contains(logs.String(), " 200 ") {
		t.Fatalf(
			"expected log to contain status 200, got %q",
			logs.String(),
		)
	}
}

func TestLoggingMiddleware_Recovery(t *testing.T) {
	logger, logs := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handler := middleware.Recovery(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		panic("test panic")
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusInternalServerError,
			recorder.Code,
		)
	}

	expectedBody := "Internal Server Error\n"
	if recorder.Body.String() != expectedBody {
		t.Fatalf(
			"expected body %q, got %q",
			expectedBody,
			recorder.Body.String(),
		)
	}

	logOutput := logs.String()

	if !strings.Contains(
		logOutput,
		"Recovered from panic: test panic",
	) {
		t.Fatalf(
			"expected recovery log, got %q",
			logOutput,
		)
	}

	if !strings.Contains(logOutput, "Stack trace:") {
		t.Fatalf(
			"expected stack trace log, got %q",
			logOutput,
		)
	}
}

func TestLoggingMiddleware_RecoveryWithoutPanic(t *testing.T) {
	logger, logs := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handlerCalled := false

	handler := middleware.Recovery(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		handlerCalled = true
		w.WriteHeader(http.StatusAccepted)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/ok",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)
	if !handlerCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusAccepted,
			recorder.Code,
		)
	}

	if strings.Contains(logs.String(), "Recovered from panic") {
		t.Fatal("did not expect recovery log")
	}
}

func TestLoggingMiddleware_CORS(t *testing.T) {
	logger, _ := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handlerCalled := false

	handler := middleware.CORS(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/posts",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !handlerCalled {
		t.Fatal("expected next handler to be called")
	}

	expectedHeaders := map[string]string{
		"Access-Control-Allow-Origin":  "*",
		"Access-Control-Allow-Methods": "GET, POST, PUT, DELETE, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type, Authorization",
		"Access-Control-Max-Age":       "86400",
	}

	for header, expectedValue := range expectedHeaders {
		actualValue := recorder.Header().Get(header)

		if actualValue != expectedValue {
			t.Errorf(
				"header %q: expected %q, got %q",
				header,
				expectedValue,
				actualValue,
			)
		}
	}
}

func TestLoggingMiddleware_CORS_Preflight(t *testing.T) {
	logger, _ := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handlerCalled := false

	handler := middleware.CORS(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		handlerCalled = true
	})

	request := httptest.NewRequest(
		http.MethodOptions,
		"/posts",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if handlerCalled {
		t.Fatal("next handler must not be called for OPTIONS request")
	}

	if recorder.Code != http.StatusNoContent {
		t.Fatalf(
			"expected status code %d, got %d",
			http.StatusNoContent,
			recorder.Code,
		)
	}

	if recorder.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("expected CORS headers to be set")
	}
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
	logger, logs := testLogger()
	middleware := NewLoggingMiddleware(logger)

	var requestIDFromContext string

	handler := middleware.RequestID(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		value := r.Context().Value("RequestID")

		requestID, ok := value.(string)
		if !ok {
			t.Fatal("expected RequestID string in context")
		}

		requestIDFromContext = requestID

		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/profile",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	requestIDFromHeader := recorder.Header().Get("X-Request-ID")

	if requestIDFromHeader == "" {
		t.Fatal("expected X-Request-ID response header")
	}

	if requestIDFromContext == "" {
		t.Fatal("expected RequestID in request context")
	}

	if requestIDFromHeader != requestIDFromContext {
		t.Fatalf(
			"header RequestID %q differs from context RequestID %q",
			requestIDFromHeader,
			requestIDFromContext,
		)
	}

	if _, err := uuid.Parse(requestIDFromHeader); err != nil {
		t.Fatalf(
			"expected valid UUID in X-Request-ID, got %q: %v",
			requestIDFromHeader,
			err,
		)
	}

	expectedLogParts := []string{
		"Received request GET /profile with Request ID:",
		requestIDFromHeader,
	}

	logOutput := logs.String()

	for _, expected := range expectedLogParts {
		if !strings.Contains(logOutput, expected) {
			t.Errorf(
				"expected log to contain %q, got %q",
				expected,
				logOutput,
			)
		}
	}
}

func TestLoggingMiddleware_RequestID_GeneratesDifferentIDs(
	t *testing.T,
) {
	logger, _ := testLogger()
	middleware := NewLoggingMiddleware(logger)

	handler := middleware.RequestID(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		w.WriteHeader(http.StatusOK)
	})

	firstRecorder := httptest.NewRecorder()
	firstRequest := httptest.NewRequest(
		http.MethodGet,
		"/first",
		nil,
	)

	handler(firstRecorder, firstRequest)
	secondRecorder := httptest.NewRecorder()
	secondRequest := httptest.NewRequest(
		http.MethodGet,
		"/second",
		nil,
	)

	handler(secondRecorder, secondRequest)

	firstID := firstRecorder.Header().Get("X-Request-ID")
	secondID := secondRecorder.Header().Get("X-Request-ID")

	if firstID == "" || secondID == "" {
		t.Fatal("expected both requests to have request IDs")
	}

	if firstID == secondID {
		t.Fatalf(
			"expected different request IDs, got %q",
			firstID,
		)
	}
}

func TestLoggingMiddleware_RequestID_PreservesRequestContext(
	t *testing.T,
) {
	logger, _ := testLogger()
	middleware := NewLoggingMiddleware(logger)

	const contextKey = "original-key"
	const contextValue = "original-value"

	handler := middleware.RequestID(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		if got := r.Context().Value(contextKey); got != contextValue {
			t.Fatalf(
				"expected original context value %q, got %v",
				contextValue,
				got,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	request := httptest.NewRequest(
		http.MethodGet,
		"/context",
		nil,
	)

	request = request.WithContext(
		context.WithValue(
			request.Context(),
			contextKey,
			contextValue,
		),
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)
}
