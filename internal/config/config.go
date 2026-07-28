package config

import (
	"log"
	"os"
	"strconv"
)

// Config представляет конфигурацию приложения
type Config struct {
	// Server
	ServerHost string
	ServerPort int

	// Database
	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	// JWT
	JWTSecret      string
	JWTExpiryHours int

	// Cache
	CacheTTLMinutes int
}

// loadConfig загружает конфигурацию из переменных окружения
// Уст. разумные значения по умолчанию
func EnvloadConfig() *Config {
	log.Printf("EnvloadConfig()")

	return &Config{
		ServerHost:      getEnv("SERVER_HOST", "localhost"),
		ServerPort:      getEnvAsInt("SERVER_PORT", 8080),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnvAsInt("DB_PORT", 5432),
		DBUser:          getEnv("DB_USER", "bloguser"),
		DBPassword:      getEnv("DB_PASSWORD", "postgres"),
		DBName:          getEnv("DB_NAME", "blogdb"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "default-secret-key"),
		JWTExpiryHours:  getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		CacheTTLMinutes: getEnvAsInt("CACHE_TTL_MINUTES", 5),
	}
}

// getEnv получает значение переменной окружения или возвращает значение по умолчанию
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt получает значение переменной окружения как int или возвращает значение по умолчанию
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}
