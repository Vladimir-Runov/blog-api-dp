package repository

import (
	blogerrors "blog-api-dp/internal/erros"
	"blog-api-dp/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// UserRepo представляет репозиторий для работы с пользователями
type UserRepo struct {
	db *sql.DB
}

// NewUserRepo создает новый репозиторий пользователей
func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

// TODO: Реализовать создание пользователя
// Create создает нового пользователя
func (r *UserRepo) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (username, email, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	now := time.Now()
	// 2. Установить created_at и updated_at = time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	//  Выполнить запрос и получить ID созданной записи
	//  Установить ID в структуру user
	err := r.db.QueryRowContext(ctx, query, user.Username, user.Email, user.Password, user.CreatedAt, user.UpdatedAt).Scan(&user.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("no rows were returned")
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID получает пользователя по ID
// TODO: Реализовать получение пользователя по ID
func (r *UserRepo) GetByID(ctx context.Context, id int) (*model.User, error) {
	query := `
		SELECT id, username, email, password, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var user model.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil { // Не забудьте обработать sql.ErrNoRows и вернуть ErrUserNotFound
		if err == sql.ErrNoRows { // Обработка случая, когда пользователь не найден
			return nil, blogerrors.ErrUserNotFound // Возвращаем sentinel error, чтобы handler мог вернуть 404 Not Found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil // Возвращаем найденного пользователя
}

// GetByEmail получает пользователя по email
// TODO: Реализовать получение пользователя по email
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {

	query := `
        SELECT id, username, email, password, created_at, updated_at
        FROM users
        WHERE email = $1
    `

	var user model.User

	err := r.db.QueryRowContext(ctx, query, email).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil { // Не забудьте обработать sql.ErrNoRows и вернуть ErrUserNotFound
		if err == sql.ErrNoRows { // Обработка случая, когда пользователь не найден
			return nil, blogerrors.ErrUserNotFound // Возвращаем sentinel error, чтобы handler мог вернуть 404 Not Found
		}
		return nil, fmt.Errorf("failed to get user by E-mail: %w", err)
	}

	return &user, nil // Возвращаем найденного пользователя
}

// GetByUsername получает пользователя по username
// TODO: Реализовать получение пользователя по username
func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	query := `
        SELECT id, username, email, password, created_at, updated_at
        FROM users
        WHERE username = $1
    `

	var user model.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(&user.ID, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, blogerrors.ErrUserNotFound // Возвращаем sentinel error, чтобы handler мог вернуть 404 Not Found
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil // Возвращаем найденного пользователя
}

// ExistsByEmail проверяет существование пользователя по email
func (r *UserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	// TODO: Реализовать проверку существования пользователя
	// HINT: Используйте SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`

	var exists bool
	// TODO: Выполнить запрос и просканировать результат в переменную exists

	//_ = query // Удалите эту строку после реализации
	//return false, fmt.Errorf("not implemented")

	err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of user: %w", err)
	}

	return exists, nil // Возвращаем результат проверки
}

// ExistsByUsername проверяет существование пользователя по username
func (r *UserRepo) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	// TODO: Реализовать проверку существования пользователя по username
	// Аналогично ExistsByEmail

	//return false, fmt.Errorf("not implemented")
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check existence of user: %w", err)
	}

	return exists, nil // Возвращаем результат проверки
}

// Update обновляет данные пользователя
func (r *UserRepo) Update(ctx context.Context, user *model.User) error {
	// TODO: (Опционально) Реализовать обновление пользователя
	// 1. Подготовить SQL запрос UPDATE users SET ... WHERE id = $X
	// 2. Обновить updated_at = time.Now()
	// 3. Выполнить запрос
	// 4. Проверить, что запись была обновлена (RowsAffected)

	//return fmt.Errorf("not implemented")
	// Подготовить SQL запрос для обновления данных пользователя

	query := `
        UPDATE users 
        SET username = $1, email = $2, updated_at = $3 
        WHERE id = $4
    `

	// Установим текущее время для updated_at
	updatedAt := time.Now()

	// Выполнить запрос
	res, err := r.db.ExecContext(ctx, query, user.Username, user.Email, updatedAt, user.ID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	// Проверить, что запись была обновлена
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows were updated")
	}

	return nil // Успешное обновление
}

// Delete удаляет пользователя
func (r *UserRepo) Delete(ctx context.Context, id int) error {
	// TODO: (Опционально) Реализовать удаление пользователя
	// 1. Подготовить SQL запрос DELETE FROM users WHERE id = $1
	// 2. Выполнить запрос
	// 3. Проверить, что запись была удалена (RowsAffected)

	//return fmt.Errorf("not implemented")
	// Подготовить SQL запрос для удаления пользователя
	query := `DELETE FROM users WHERE id = $1`

	// Выполнить запрос
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Проверить, что запись была удалена
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no rows were deleted")
	}

	return nil // Успешное удаление
}
