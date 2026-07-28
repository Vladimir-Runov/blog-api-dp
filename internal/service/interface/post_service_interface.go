package service

import (
	"blog-api-dp/internal/model"
	"context"
)

// PostServiceInterface определяет интерфейс для сервиса постов.
// Используется для мока в юнит-тестах.
type PostServiceInterface interface {
	// Create создаёт новый пост
	Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, error)

	// GetByID получает пост по ID с учётом прав доступа.
	GetByID(ctx context.Context, id int, requestorID int) (*model.Post, error)

	// GetAll получает все опубликованные посты с пагинацией
	GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error)

	// Update обновляет пост (только автор)
	Update(ctx context.Context, id int, userID int, req *model.PostUpdateRequest) (*model.Post, error)

	// Delete удаляет пост (только автор)
	Delete(ctx context.Context, id int, userID int) error

	// GetByAuthor получает посты автора с пагинацией
	GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, int, error)
}
