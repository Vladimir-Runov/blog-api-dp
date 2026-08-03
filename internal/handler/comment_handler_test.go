package handler

// go test ./internal/handler/...

//тесты:
//• Create с неправильным HTTP-методом;
//• Create без пользователя в контексте;
//• GetByID с неправильным HTTP-методом;
//• GetByID с некорректным ID;
//• GetByID без ID в URL.

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCommentHandler_Create_MethodNotAllowed(t *testing.T) {
	handler := NewCommentHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/posts/1/comments",
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

func TestCommentHandler_Create_Unauthorized(t *testing.T) {
	handler := NewCommentHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/posts/1/comments",
		strings.NewReader(`{"content":"test comment"}`),
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

func TestCommentHandler_GetByID_MethodNotAllowed(t *testing.T) {
	handler := NewCommentHandler(nil)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/comments/1",
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

func TestCommentHandler_GetByID_InvalidID(t *testing.T) {
	handler := NewCommentHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/comments/not-a-number",
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

func TestCommentHandler_GetByID_MissingID(t *testing.T) {
	handler := NewCommentHandler(nil)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/comments/",
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
