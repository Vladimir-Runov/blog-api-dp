package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPostHandler_Create_MethodNotAllowed(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/posts",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestPostHandler_Create_Unauthorized(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/posts",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.Create(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			rec.Code,
		)
	}
}

func TestPostHandler_GetByID_MethodNotAllowed(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/posts/1",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}

func TestPostHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/posts/not-a-number",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestPostHandler_GetByID_MissingID(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/posts/",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusBadRequest,
			rec.Code,
		)
	}
}

func TestPostHandler_GetAll_MethodNotAllowed(t *testing.T) {
	handler := NewPostHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/posts",
		nil,
	)
	rec := httptest.NewRecorder()

	handler.GetAll(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusMethodNotAllowed,
			rec.Code,
		)
	}
}
