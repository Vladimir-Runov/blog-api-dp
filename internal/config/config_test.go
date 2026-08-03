package config

import (
	"testing"
)

func TestEnvloadConfig_DefaultValues(t *testing.T) {

	t.Setenv("SERVER_HOST", "")
	t.Setenv("SERVER_PORT", "")
	t.Setenv("DB_HOST", "")
	t.Setenv("DB_PORT", "")
	t.Setenv("DB_USER", "")
	t.Setenv("DB_PASSWORD", "")
	t.Setenv("DB_NAME", "")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("JWT_SECRET", "")
	t.Setenv("JWT_EXPIRY_HOURS", "")
	t.Setenv("CACHE_TTL_MINUTES", "")

	cfg := EnvloadConfig()

	expected := &Config{
		ServerHost:      "localhost",
		ServerPort:      8080,
		DBHost:          "localhost",
		DBPort:          5433,
		DBUser:          "bloguser",
		DBPassword:      "postgres",
		DBName:          "blogdb",
		DBSSLMode:       "disable",
		JWTSecret:       "default-secret-key",
		JWTExpiryHours:  24,
		CacheTTLMinutes: 5,
	}

	if *cfg != *expected {
		t.Errorf("expected config %+v, got %+v", *expected, *cfg)
	}
}

func TestEnvloadConfig_CustomValues(t *testing.T) {
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("DB_HOST", "db.example.com")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_USER", "testuser")
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("DB_NAME", "testdb")
	t.Setenv("DB_SSLMODE", "require")
	t.Setenv("JWT_SECRET", "my-secret")
	t.Setenv("JWT_EXPIRY_HOURS", "72")
	t.Setenv("CACHE_TTL_MINUTES", "15")

	cfg := EnvloadConfig()

	expected := &Config{
		ServerHost:      "127.0.0.1",
		ServerPort:      9090,
		DBHost:          "db.example.com",
		DBPort:          5432,
		DBUser:          "testuser",
		DBPassword:      "testpassword",
		DBName:          "testdb",
		DBSSLMode:       "require",
		JWTSecret:       "my-secret",
		JWTExpiryHours:  72,
		CacheTTLMinutes: 15,
	}

	if *cfg != *expected {
		t.Errorf("expected config %+v, got %+v", *expected, *cfg)
	}
}

func TestEnvloadConfig_InvalidIntegerValues(t *testing.T) {
	t.Setenv("SERVER_PORT", "invalid")
	t.Setenv("DB_PORT", "5432abc")
	t.Setenv("JWT_EXPIRY_HOURS", "")
	t.Setenv("CACHE_TTL_MINUTES", "1.5")

	cfg := EnvloadConfig()

	if cfg.ServerPort != 8080 {
		t.Errorf("expected ServerPort to be 8080, got %d", cfg.ServerPort)
	}

	if cfg.DBPort != 5433 {
		t.Errorf("expected DBPort to be 5433, got %d", cfg.DBPort)
	}

	if cfg.JWTExpiryHours != 24 {
		t.Errorf("expected JWTExpiryHours to be 24, got %d", cfg.JWTExpiryHours)
	}

	if cfg.CacheTTLMinutes != 5 {
		t.Errorf("expected CacheTTLMinutes to be 5, got %d", cfg.CacheTTLMinutes)
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		value        string
		defaultValue string
		expected     string
	}{
		{
			name:         "returns environment value",
			key:          "TEST_CONFIG_VALUE",
			value:        "custom-value",
			defaultValue: "default-value",
			expected:     "custom-value",
		},
		{
			name:         "returns default for empty value",
			key:          "TEST_CONFIG_EMPTY",
			value:        "",
			defaultValue: "default-value",
			expected:     "default-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)

			got := getEnv(tt.key, tt.defaultValue)

			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		value        string
		defaultValue int
		expected     int
	}{
		{
			name:         "returns valid integer",
			value:        "1234",
			defaultValue: 10,
			expected:     1234,
		},
		{
			name:         "returns default for invalid integer",
			value:        "invalid",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "returns default for empty value",
			value:        "",
			defaultValue: 10,
			expected:     10,
		},
		{
			name:         "supports negative integer",
			value:        "-5",
			defaultValue: 10,
			expected:     -5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("TEST_CONFIG_INT", tt.value)

			got := getEnvAsInt("TEST_CONFIG_INT", tt.defaultValue)

			if got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}
