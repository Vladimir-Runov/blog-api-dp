package auth

// go test ./auth
// go test ./auth -cover

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	testSecret   = "test-secret_key-here-0000000000000000000000000000000000000000000000000000000"
	testUserID   = 42
	testEmail    = "user@example.com"
	testUsername = "testuser"
)

func TestNewJWTManager(t *testing.T) {
	manager := NewJWTManager(testSecret, 2)

	if manager == nil {
		t.Fatal("expected manager, got nil")
	}

	if string(manager.secretKey) != testSecret {
		t.Errorf(
			"expected secret %q, got %q",
			testSecret,
			string(manager.secretKey),
		)
	}

	expectedTTL := 2 * time.Hour

	if manager.ttl != expectedTTL {
		t.Errorf(
			"expected TTL %v, got %v",
			expectedTTL,
			manager.ttl,
		)
	}
}

func TestGenerateToken(t *testing.T) {
	manager := NewJWTManager(testSecret, 1)

	tokenString, expirationTime, err := manager.GenerateToken(
		testUserID,
		testEmail,
		testUsername,
	)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	if tokenString == "" {
		t.Fatal("expected non-empty token")
	}

	if expirationTime.Before(time.Now()) {
		t.Fatal("expected expiration time to be in the future")
	}

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(testSecret), nil
		},
	)
	if err != nil {
		t.Fatalf("failed to parse generated token: %v", err)
	}

	if !token.Valid {
		t.Fatal("expected generated token to be valid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		t.Fatal("expected token claims to have type *Claims")
	}

	if claims.UserID != testUserID {
		t.Errorf(
			"expected user ID %d, got %d",
			testUserID,
			claims.UserID,
		)
	}

	if claims.Email != testEmail {
		t.Errorf(
			"expected email %q, got %q",
			testEmail,
			claims.Email,
		)
	}

	if claims.Username != testUsername {
		t.Errorf(
			"expected username %q, got %q",
			testUsername,
			claims.Username,
		)
	}

	if claims.ExpiresAt == nil {
		t.Fatal("expected ExpiresAt claim")
	}

	if claims.IssuedAt == nil {
		t.Fatal("expected IssuedAt claim")
	}
}

func TestValidateToken_ValidToken(t *testing.T) {
	manager := NewJWTManager(testSecret, 1)

	tokenString, _, err := manager.GenerateToken(
		testUserID,
		testEmail,
		testUsername,
	)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := manager.ValidateToken(tokenString)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims == nil {
		t.Fatal("expected claims, got nil")
	}

	if claims.UserID != testUserID {
		t.Errorf(
			"expected user ID %d, got %d",
			testUserID,
			claims.UserID,
		)
	}

	if claims.Email != testEmail {
		t.Errorf(
			"expected email %q, got %q",
			testEmail,
			claims.Email,
		)
	}

	if claims.Username != testUsername {
		t.Errorf(
			"expected username %q, got %q",
			testUsername,
			claims.Username,
		)
	}
}

func TestValidateToken_InvalidSignature(t *testing.T) {
	manager := NewJWTManager(testSecret, 1)
	anotherManager := NewJWTManager("another-secret", 1)

	tokenString, _, err := anotherManager.GenerateToken(
		testUserID,
		testEmail,
		testUsername,
	)
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}

	claims, err := manager.ValidateToken(tokenString)

	if claims != nil {
		t.Fatalf("expected nil claims, got %+v", claims)
	}

	if !errors.Is(err, ErrInvalidToken) {
		t.Fatalf(
			"expected ErrInvalidToken, got %v",
			err,
		)
	}
}

func TestValidateToken_MalformedToken(t *testing.T) {
	manager := NewJWTManager(testSecret, 1)

	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "random string",
			token: "not-a-jwt-token",
		},
		{
			name:  "invalid jwt structure",
			token: "header.payload",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := manager.ValidateToken(tt.token)

			if claims != nil {
				t.Fatalf("expected nil claims, got %+v", claims)
			}

			if !errors.Is(err, ErrInvalidToken) {
				t.Fatalf(
					"expected ErrInvalidToken, got %v",
					err,
				)
			}
		})
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {

	manager := NewJWTManager(testSecret, 1)

	expirationTime := time.Now().Add(-time.Hour * 100) // Установить время истечения токена в прошлом (например, 100 часов назад)

	expiredClaims := Claims{
		UserID:   testUserID,
		Email:    testEmail,
		Username: testUsername,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(expirationTime.Add(-time.Hour)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		expiredClaims,
	)

	tokenString, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	claims, err := manager.ValidateToken(tokenString)

	if claims != nil {
		t.Fatalf("expected nil claims, got %+v", claims)
	}

	if !errors.Is(err, ErrExpiredToken) {
		t.Fatalf(
			"expected ErrExpiredToken, got %v",
			err,
		)
	}
}
