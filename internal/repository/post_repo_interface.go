package repository

import (
	"blog-api-dp/internal/model"
	"context"
)

// PostRepository определяет интерфейс для работы с постами
type PostRepository interface {
	// Create создает новый пост
	Create(ctx context.Context, post *model.Post) error

	// GetByID получает пост по ID
	GetByID(ctx context.Context, id int) (*model.Post, error)

	// GetAll получает все посты с пагинацией
	// limit - количество записей на странице
	// offset - смещение от начала
	GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error)

	// GetTotalCount получает общее количество постов
	GetTotalCount(ctx context.Context) (int, error)

	// Update обновляет пост
	Update(ctx context.Context, post *model.Post) error

	// Delete удаляет пост по ID
	Delete(ctx context.Context, id int) error

	// Exists проверяет существование поста по ID
	Exists(ctx context.Context, id int) (bool, error)

	// TODO: Добавить методы для получения постов конкретного автора
	GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error)

	// GetTotalCountByAuthorID получает общее количество постов автора
	CountByAuthorID(ctx context.Context, authorID int) (int, error)

	GetScheduledPosts(ctx context.Context) ([]*model.Post, error)

	PublishPost(ctx context.Context, id int) error
}
