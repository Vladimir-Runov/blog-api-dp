package handler

// go test ./internal/handler/...

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAuthHandler_Register_MethodNotAllowed(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/register",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/register",
		strings.NewReader(`{"username":`),
	)
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestAuthHandler_Login_MethodNotAllowed(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/login",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{"email":`),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestAuthHandler_Login_InvalidRequest(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{}`),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestAuthHandler_Login_InvalidRequestData(t *testing.T) {
	handler := NewAuthHandlerEx(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/login",
		strings.NewReader(`{
   "email": "",
   "password": ""
  }`),
	)
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}
