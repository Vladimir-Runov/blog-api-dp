package main

import (
	"blog-example-go-restapi/internal/handler"
	middlewareauth "blog-example-go-restapi/internal/middleware"
	"blog-example-go-restapi/internal/repository"
	"blog-example-go-restapi/internal/service"
	"blog-example-go-restapi/pkg/auth"
	"blog-example-go-restapi/pkg/database"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем конфигурацию из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using environment variables")
	}

	//  Загрузить конфигурацию из переменных окружения
	cfg := loadConfig()

	//  Подключиться к базе данных
	// - Создать database.Config из параметров конфигурации
	dbConfig := database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	}


}

// loadConfig загружает конфигурацию из переменных окружения
func loadConfig() *Config {
	//  Реализовать загрузку всех параметров конфигурации
	// Использовать вспомогательные функции getEnv и getEnvAsInt
	// Установить разумные значения по умолчанию
	return &Config{
		ServerHost:      getEnv("SERVER_HOST", "localhost"),
		ServerPort:      getEnvAsInt("SERVER_PORT", 8080),
		DBHost:          getEnv("DB_HOST", "localhost"),
		DBPort:          getEnvAsInt("DB_PORT", 5432),
		DBUser:          getEnv("DB_USER", "user"),
		DBPassword:      getEnv("DB_PASSWORD", "password"),
		DBName:          getEnv("DB_NAME", "dbname"),
		DBSSLMode:       getEnv("DB_SSLMODE", "disable"),
		JWTSecret:       getEnv("JWT_SECRET", "supersecretkey"),
		JWTExpiryHours:  getEnvAsInt("JWT_EXPIRY_HOURS", 24),
		CacheTTLMinutes: getEnvAsInt("CACHE_TTL_MINUTES", 60),
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
