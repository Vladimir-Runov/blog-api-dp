package service

import (
	"blog-api-dp/internal/model"
	"context"
)

// CommentServiceInterface определяет интерфейс для сервиса комментариев.
// Используется для мокирования в юнит-тестах.
type CommentServiceInterface interface {
	// Create создаёт новый комментарий.
	Create(ctx context.Context, userID, postID int, content string) (*model.Comment, error)

	// GetByPost получает комментарии к посту с пагинацией.
	GetByPost(ctx context.Context, postID, limit, offset int) ([]*model.Comment, int, error)
}
