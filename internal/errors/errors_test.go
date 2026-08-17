package blogerrors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestApplicationErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{
			name: "ErrInvalidCredentials",
			err:  ErrInvalidCredentials,
			msg:  "invalid credentials",
		},
		{
			name: "ErrUserAlreadyExists",
			err:  ErrUserAlreadyExists,
			msg:  "user already exists",
		},
		{
			name: "ErrUserNotFound",
			err:  ErrUserNotFound,
			msg:  "user not found",
		},
		{
			name: "ErrInvalidPostID",
			err:  ErrInvalidPostID,
			msg:  "invalid post ID",
		},
		{
			name: "ErrPostNotFound",
			err:  ErrPostNotFound,
			msg:  "post not found",
		},
		{
			name: "ErrPostNotExists",
			err:  ErrPostNotExists,
			msg:  "post does not exist",
		},
		{
			name: "ErrUnauthorized",
			err:  ErrUnauthorized,
			msg:  "unauthorized",
		},
		{
			name: "ErrForbidden",
			err:  ErrForbidden,
			msg:  "forbidden",
		},
		{
			name: "ErrCommentNotFound",
			err:  ErrCommentNotFound,
			msg:  "comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatal("error is nil")
			}

			if got := tt.err.Error(); got != tt.msg {
				t.Errorf("expected %q, got %q", tt.msg, got)
			}
		})
	}
}

func TestReplyJsonError(t *testing.T) {
	recorder := httptest.NewRecorder()

	statusCode := http.StatusNotFound
	message := "post not found"

	ReplyJsonError(recorder, message, statusCode)

	if recorder.Code != statusCode {
		t.Errorf("expected status code %d, got %d", statusCode, recorder.Code)
	}

	expectedContentType := "application/json"
	if got := recorder.Header().Get("Content-Type"); got != expectedContentType {
		t.Errorf(
			"expected Content-Type %q, got %q",
			expectedContentType,
			got,
		)
	}

	var response ErrorResponse

	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}

	if response.Error != http.StatusText(statusCode) {
		t.Errorf(
			"expected error %q, got %q",
			http.StatusText(statusCode),
			response.Error,
		)
	}

	if response.Message != message {
		t.Errorf(
			"expected message %q, got %q",
			message,
			response.Message,
		)
	}
}
