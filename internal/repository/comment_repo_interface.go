package repository

import (
	"blog-api-dp/internal/model"
	"context"
)

// CommentRepository определяет интерфейс для работы с комментариями
type CommentRepository interface {
	// Create создает новый комментарий
	Create(ctx context.Context, comment *model.Comment) error

	// GetByID получает комментарий по ID
	GetByID(ctx context.Context, id int) (*model.Comment, error)

	// GetByPostID получает комментарии к посту с пагинацией
	GetByPostID(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, error)

	// GetCountByPostID получает количество комментариев к посту
	GetCountByPostID(ctx context.Context, postID int) (int, error)

	Delete(ctx context.Context, id int) error

	Update(ctx context.Context, comment *model.Comment) error
}
