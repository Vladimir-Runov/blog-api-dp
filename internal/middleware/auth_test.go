package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"blog-api-dp/pkg/auth"
)

const (
	testJWTSecret = "test-secret-key-here-0000000000000000000000000000000000000000000000000000000"
	testUserID    = 42
	testEmail     = "user@example.com"
	timeHour      = 1
	testUsername  = "some_test_user"
)

func newTestJWTManager() *auth.JWTManager {
	return auth.NewJWTManager(testJWTSecret, timeHour)
}

func newTestToken(t *testing.T) string {
	t.Helper()

	manager := newTestJWTManager()

	token, _, err := manager.GenerateToken(testUserID, testEmail, testUsername)
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}

	return token
}

func TestAuth_ExtractToken(t *testing.T) {
	tests := []struct {
		name          string
		authorization string
		expected      string
	}{
		{
			name:          "valid bearer token",
			authorization: "Bearer abc.def.ghi",
			expected:      "abc.def.ghi",
		},
		{
			name:          "missing authorization header",
			authorization: "",
			expected:      "",
		},
		{
			name:          "wrong authentication scheme",
			authorization: "Basic abc.def.ghi",
			expected:      "",
		},
		{
			name:          "only bearer scheme",
			authorization: "Bearer",
			expected:      "",
		},
		{
			name:          "token without bearer scheme",
			authorization: "abc.def.ghi",
			expected:      "",
		},
		{
			name:          "too many parts",
			authorization: "Bearer abc def",
			expected:      "",
		},
		{
			name:          "lowercase bearer",
			authorization: "bearer abc.def.ghi",
			expected:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(
				http.MethodGet,
				"/",
				nil,
			)

			request.Header.Set(
				"Authorization",
				tt.authorization,
			)

			got := extractToken(request)

			if got != tt.expected {
				t.Fatalf(
					"expected token %q, got %q",
					tt.expected,
					got,
				)
			}
		})
	}
}

func TestAuth_RequireAuth_MissingToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true
	})

	handler := middleware.RequireAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/private",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if nextCalled {
		t.Fatal("next handler must not be called")
	}

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	expectedMessage := "Unauthorized...Missing token"

	if !strings.Contains(
		recorder.Body.String(),
		expectedMessage,
	) {
		t.Fatalf(
			"expected response to contain %q, got %q",
			expectedMessage,
			recorder.Body.String(),
		)
	}
}

func TestAuth_RequireAuth_InvalidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true
	})

	handler := middleware.RequireAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/private",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer invalid-token",
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if nextCalled {
		t.Fatal("next handler must not be called")
	}

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusUnauthorized,
			recorder.Code,
		)
	}

	if !strings.Contains(
		recorder.Body.String(),
		"Unauthorized...Invalid token",
	) {
		t.Fatalf(
			"expected invalid token response, got %q",
			recorder.Body.String(),
		)
	}
}

func TestAuth_RequireAuth_ValidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true

		userID := r.Context().Value(UserIDKey)
		userEmail := r.Context().Value(UserEmailKey)

		username := r.Context().Value(UserNameKey)

		if userID == nil {
			t.Error("expected user ID in context")
		}

		if userEmail != testEmail {
			t.Errorf(
				"expected email %q, got %v",
				testEmail,
				userEmail,
			)
		}

		if username != testUsername {
			t.Errorf(
				"expected username %q, got %v",
				testUsername,
				username,
			)
		}

		if got := userIDToString(userID); got != "42" {
			t.Errorf(
				"expected user ID %q, got %q",
				"42",
				got,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RequireAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/private",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer "+newTestToken(t),
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func TestOptionalAuth_WithoutToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true

		if value := r.Context().Value(UserIDKey); value != nil {
			t.Errorf(
				"did not expect user ID in context, got %v",
				value,
			)
		}

		if value := r.Context().Value(UserEmailKey); value != nil {
			t.Errorf(
				"did not expect email in context, got %v",
				value,
			)
		}

		if value := r.Context().Value(UserNameKey); value != nil {
			t.Errorf(
				"did not expect username in context, got %v",
				value,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/public",
		nil,
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func TestOptionalAuth_WithInvalidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true

		if value := r.Context().Value(UserIDKey); value != nil {
			t.Errorf(
				"did not expect user ID in context, got %v",
				value,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/public",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer invalid-token",
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(
			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func TestOptionalAuth_WithValidToken(t *testing.T) {
	jwtManager := newTestJWTManager()
	middleware := NewAuthMiddleware(jwtManager)

	nextCalled := false

	next := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		nextCalled = true

		userID := r.Context().Value(UserIDKey)
		userEmail := r.Context().Value(UserEmailKey)
		username := r.Context().Value(UserNameKey)

		if userID == nil {
			t.Error("expected user ID in context")
		}

		if userEmail != testEmail {
			t.Errorf(
				"expected email %q, got %v",
				testEmail,
				userEmail,
			)
		}

		if username != testUsername {
			t.Errorf(
				"expected username %q, got %v",
				testUsername,
				username,
			)
		}

		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.OptionalAuth(next)

	request := httptest.NewRequest(
		http.MethodGet,
		"/public",
		nil,
	)

	request.Header.Set(
		"Authorization",
		"Bearer "+newTestToken(t),
	)

	recorder := httptest.NewRecorder()

	handler(recorder, request)

	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf(

			"expected status %d, got %d",
			http.StatusOK,
			recorder.Code,
		)
	}
}

func userIDToString_del(value any) string {
	switch userID := value.(type) {
	case int:
		return string(rune('0' + userID))
	case int64:
		return string(rune('0' + userID))
	case string:
		return userID
	default:
		return ""
	}
}
func userIDToString(value any) string {
	return fmt.Sprint(value)
}
