package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/lib/pq"
)

// = "migrations"
func getMigrationQueries(migrationsDir string) ([]string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Printf("Failed to get current directory: %v", err)
	} else {
		log.Printf("Looking for migrations in dif: %s  %s", cwd, migrationsDir)
	}

	log.Printf(".. Looking for migrations in: %s", migrationsDir)

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read migrations directory: %w", err)
	}
	var queries []string // Объявляем срез без фиксированного размера
	var migrationFiles []string
	migrationPattern := regexp.MustCompile(`^\d+_.+\.sql$`)

	for _, file := range files {
		if !file.IsDir() && migrationPattern.MatchString(file.Name()) {
			migrationFiles = append(migrationFiles, file.Name())
		}
	}

	// Сортируем миграции по номеру
	sort.Slice(migrationFiles, func(i, j int) bool {
		return extractMigrationNumber(migrationFiles[i]) < extractMigrationNumber(migrationFiles[j])
	})

	// Проверяем, нужно ли применять миграции
	for _, file := range migrationFiles {
		//log.Printf("Processing migration: %s", file)

		content, err := os.ReadFile(filepath.Join(migrationsDir, file))
		if err != nil {
			return nil, fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		// Разбиваем SQL на отдельные запросы
		queriesDx := splitSQLQueries(string(content))
		for _, query := range queriesDx {
			//fmt.Println(query)
			queries = append(queries, query)
		}
	}
	return queries, nil
}

// extractMigrationNumber извлекает номер миграции из имени файла
func extractMigrationNumber(filename string) int {
	re := regexp.MustCompile(`^(\d+)_`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) > 1 {
		var num int
		fmt.Sscanf(matches[1], "%d", &num)
		return num
	}
	return -1
}

// splitSQLQueries разбивает SQL-скрипт на отдельные запросы
func splitSQLQueries(content string) []string {
	var queries []string
	var currentQuery strings.Builder
	inString := false
	stringChar := byte(0)
	inComment := false
	commentType := "" // "--" или "/*"

	for i := 0; i < len(content); i++ {
		char := content[i]

		// Обработка комментариев
		if !inString && !inComment && i+1 < len(content) {
			if content[i:i+2] == "--" {
				inComment = true
				commentType = "--"
				i++ // Пропускаем следующий символ
				continue
			} else if content[i:i+2] == "/*" {
				inComment = true
				commentType = "/*"
				i++ // Пропускаем следующий символ
				continue
			}
		}

		// Выход из комментариев
		if inComment {
			if commentType == "--" && char == '\n' {
				inComment = false
			} else if commentType == "/*" && i+1 < len(content) && content[i:i+2] == "*/" {
				inComment = false
				i++ // Пропускаем следующий символ
			}
			continue
		}

		// Обработка строк
		if char == '\'' || char == '"' {
			if !inString {
				inString = true
				stringChar = char
			} else if stringChar == char {
				// Проверяем, не экранирована ли кавычка
				if i > 0 && content[i-1] != '\\' {
					inString = false
				}
			}
		}

		// Разделение запросов
		if char == ';' && !inString {
			query := strings.TrimSpace(currentQuery.String())
			if query != "" {
				queries = append(queries, query)
			}
			currentQuery.Reset()
		} else {
			currentQuery.WriteByte(char)
		}
	}

	// Добавляем последний запрос, если он есть
	finalQuery := strings.TrimSpace(currentQuery.String())
	if finalQuery != "" {
		queries = append(queries, finalQuery)
	}

	return queries
}

// Migrate выполняет миграции базы данных
func Migrate(db *sql.DB, reset_db bool) error {
	if reset_db { // public
		_, err := db.Exec("DROP SCHEMA IF EXISTS public CASCADE; CREATE SCHEMA public;")
		if err != nil {
			return fmt.Errorf("failed to reset database: %w", err)
		}
	}
	queries, err := getMigrationQueries("migrations")
	if err != nil {
		return fmt.Errorf("failed to read queries : %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Выполняем все запросы
	for _, query := range queries {
		if _, err := tx.Exec(query); err != nil {
			tx.Rollback()
			return fmt.Errorf("error executing query: %s, error: %w", query, err)
		}

	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func getDefQueries() []string {
	queries := []string{
		// `CREATE TABLE IF NOT EXISTS users (...)`,
		`CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				username VARCHAR(50) NOT NULL UNIQUE,
				email VARCHAR(255) NOT NULL UNIQUE CHECK (email LIKE '%_@__%.__%'),
				password VARCHAR(255) NOT NULL CHECK (LENGTH(password) >= 8),
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`,

		// `CREATE TABLE IF NOT EXISTS posts (...)`,
		// `CREATE TABLE IF NOT EXISTS posts (
		//									id SERIAL PRIMARY KEY,
		//									title VARCHAR(200) NOT NULL,
		//									content TEXT NOT NULL,
		//									author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		//									created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		//									updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		//);`,
		`CREATE TABLE IF NOT EXISTS posts (
				id SERIAL PRIMARY KEY,
				title VARCHAR(200) NOT NULL CHECK (LENGTH(title) > 0),
				content TEXT NOT NULL	CHECK (LENGTH(content) > 0),
				author_id INTEGER NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				CONSTRAINT fk_posts_author 
						FOREIGN KEY (author_id) 
						REFERENCES users(id) 
						ON DELETE CASCADE
		);`,
		// `CREATE TABLE IF NOT EXISTS comments (...)`,
		//`CREATE TABLE IF NOT EXISTS comments (
		//										id SERIAL PRIMARY KEY,
		//										content TEXT NOT NULL,
		//										post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
		//										author_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		//										created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		//										updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		//);`,

		`CREATE TABLE IF NOT EXISTS comments (
				id SERIAL PRIMARY KEY,
				content TEXT NOT NULL CHECK (LENGTH(content) > 0),
				post_id INTEGER NOT NULL,
				author_id INTEGER NOT NULL,
				created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

				CONSTRAINT fk_comments_post 
					FOREIGN KEY (post_id) 
					REFERENCES posts(id) 
					ON DELETE CASCADE,

				CONSTRAINT fk_comments_author 
					FOREIGN KEY (author_id) 
					REFERENCES users(id) 
					ON DELETE CASCADE
		);`,

		// `CREATE INDEX IF NOT EXISTS ...`, индекс на user_id в таблице posts
		`CREATE INDEX IF NOT EXISTS idx_posts_author_id ON posts(author_id);`,
		`CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);`,
		`CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);`,
	}
	return queries
}
