package repository

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// CommentRepo представляет репозиторий для работы с комментариями
type CommentRepo struct {
	db *sql.DB
}

// NewCommentRepo создает новый репозиторий комментариев
func NewCommentRepo(db *sql.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

// Create создает новый комментарий
func (r *CommentRepo) Create(ctx context.Context, comment *model.Comment) error {

	query := `INSERT INTO comments (post_id, author_id, content, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	now := time.Now()
	comment.CreatedAt = now
	comment.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,

		comment.PostID,
		comment.AuthorID,
		comment.Content,
		comment.CreatedAt,
		comment.UpdatedAt,
	).Scan(&comment.ID)

	if err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	return nil
}

// GetByID получает комментарий по ID
func (r *CommentRepo) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	query := `
		SELECT id, post_id, author_id, content, created_at, updated_at FROM comments WHERE id = $1
	`
	var comment model.Comment
	// Выполнить запрос и просканировать результат
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.AuthorID,
		&comment.Content,
		&comment.CreatedAt,
		&comment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, blogerrors.ErrCommentNotFound
		}

		return nil, fmt.Errorf("failed to get comment: %w", err)
	}

	return &comment, nil
}

// GetByPostID получает комментарии к посту с пагинацией
func (r *CommentRepo) GetByPostID(ctx context.Context, postID int, limit int, offset int) ([]*model.Comment, error) {
	query := `
		SELECT id, content, post_id, author_id, created_at, updated_at
		FROM comments
		WHERE post_id = $1
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}
	defer rows.Close()

	// Итерировать по результатам
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.PostID,
			&comment.AuthorID,
			&comment.CreatedAt,
			&comment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		comments = append(comments, &comment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate comments: %w", err)
	}

	return comments, nil
}

// GetCountByPostID получает количество комментариев к посту
func (r *CommentRepo) GetCountByPostID(ctx context.Context, postID int) (int, error) {

	query := `SELECT COUNT(*) FROM comments WHERE post_id = $1`

	var count int

	err := r.db.QueryRowContext(ctx, query, postID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get comment count: %w", err)
	}

	return count, nil
}

// (Опционально) Реализовать обновление комментария. Обновить только content и updated_at
// Update обновляет комментарий
func (r *CommentRepo) Update(ctx context.Context, comment *model.Comment) error {

	query := `
		UPDATE comments SET content = $1, updated_at = $2 WHERE id = $3 AND author_id = $4
	`

	comment.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query, comment.Content, comment.UpdatedAt, comment.ID, comment.AuthorID)
	if err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("no rows updated, comment with id %d not found", comment.ID)
	}

	return nil
}

// Delete удаляет комментарий
func (r *CommentRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM comments WHERE id = $1` // AND author_id = $2

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete comment: %w", err)
	}

	// Проверка количества затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return blogerrors.ErrCommentNotFound //fmt.Errorf("no rows deleted, comment with id %d not found", id)
	}

	return nil // fmt.Errorf("not implemented")
}
