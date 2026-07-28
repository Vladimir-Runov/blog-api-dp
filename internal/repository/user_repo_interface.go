package repository

import (
	"blog-api-dp/internal/model"
	"context"
)

// UserRepository определяет интерфейс для работы с пользователями
type UserRepository interface {
	// Create создает нового пользователя в базе данных
	Create(ctx context.Context, user *model.User) error

	// GetByID получает пользователя по ID
	GetByID(ctx context.Context, id int) (*model.User, error)

	// GetByEmail получает пользователя по email
	GetByEmail(ctx context.Context, email string) (*model.User, error)

	// GetByUsername получает пользователя по username
	GetByUsername(ctx context.Context, username string) (*model.User, error)

	// ExistsByEmail проверяет существование пользователя по email
	ExistsByEmail(ctx context.Context, email string) (bool, error)

	// ExistsByUsername проверяет существование пользователя по username
	ExistsByUsername(ctx context.Context, username string) (bool, error)

	// Update обновляет данные пользователя
	//	Update(ctx context.Context, user *model.User) error

	// Delete удаляет пользователя
	//	Delete(ctx context.Context, id int) error

}
