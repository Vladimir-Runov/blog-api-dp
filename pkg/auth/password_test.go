package auth

import (
	"testing"
)

func TestGenerateRandomPassword(t *testing.T) {
	length := 12
	password, err := GenerateRandomPassword(length)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(password) != length {
		t.Fatalf("expected password length %d, got %d", length, len(password))
	}

	hasUpper := false
	hasDigit := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		}
		if char >= '0' && char <= '9' {
			hasDigit = true
		}
	}

	if !hasUpper {
		t.Fatal("expected at least one uppercase letter in the password")
	}

	if !hasDigit {
		t.Fatal("expected at least one digit in the password")
	}
}

/*
	log.Println("TestGenerateRandomPassword ... ")
	tests := []struct {
		name    string
		length     int
		wantErr bool
	}{
		{"Origin HTTP URL (Ok)", 3, false},
		{"Origin HTTP URL (Ok)", 13, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {		})
	}
*/
