package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// Config содержит параметры подключения к PostgreSQL
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// TODO: Реализовать подключение к PostgreSQL
// Шаги:
// 1. Сформировать строку подключения (DSN) из параметров конфигурации
// 2. Открыть соединение с БД используя sql.Open("postgres", dsn)
// 3. Проверить соединение методом Ping()
// 4. Настроить пул соединений (SetMaxOpenConns, SetMaxIdleConns)
// 5. Вернуть подключение или ошибку
//return nil, fmt.Errorf("not implemented")

// NewPostgresDB создает новое подключение к PostgreSQL
func NewPostgresDB(cfg Config) (*sql.DB, error) {
	// Формируем DSN (строку подключения)

	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)
	log.Print(dsn)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database open - Ok., error pinging database: %w", err)
	}

	// Настраиваем пул соединений
	//db.SetMaxOpenConns(cfg.MaxOpenConns)
	//db.SetMaxIdleConns(cfg.MaxIdleConns)

	return db, nil
}

// CheckConnection проверяет соединение с базой данных
func CheckConnection(db *sql.DB) error {
	// TODO: Реализовать проверку соединения
	// Использовать db.Ping() для проверки
	//return fmt.Errorf("not implemented")
	// Пингуем базу данных, чтобы проверить соединение
	if err := db.Ping(); err != nil {
		return fmt.Errorf("нет соединения с базой данных: %w", err)
	}
	return nil
}

// GetDSN формирует строку подключения к PostgreSQL
func GetDSN(cfg Config) string {
	// TODO: Сформировать DSN строку
	// Формат: "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s"
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)
}

// Close закрывает соединение с базой данных
func Close(db *sql.DB) error {
	// TODO: Корректно закрыть соединение
	// return fmt.Errorf("not implemented")
	if err := db.Close(); err != nil {
		return fmt.Errorf("не удалось закрыть соединение: %w", err)
	}
	return nil
}

// TestConnection выполняет тестовый запрос к БД (опциональное задание)
func TestConnection(db *sql.DB) error {
	// TODO: Выполнить простой запрос для проверки работы БД
	// Например: SELECT 1
	//return fmt.Errorf("not implemented")
	var result int
	err := db.QueryRow("SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("тестовая команда не выполнена: %w", err)
	}
	return nil
}
