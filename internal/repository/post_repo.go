package repository

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PostRepo представляет репозиторий для работы с постами
type PostRepo struct {
	db *sql.DB
}

// NewPostRepo создает новый репозиторий постов
func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

// Create создает новый пост
func (r *PostRepo) Create(ctx context.Context, post *model.Post) error {
	now := time.Now()
	post.CreatedAt = now
	post.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, PostRepo_createPostQuery,
		post.Title,
		post.Content,
		post.AuthorID,
		post.Status,
		post.PublishAt,
		post.CreatedAt,
		post.UpdatedAt).Scan(&post.ID)
	if err != nil {
		//log.Printf("QueryRowContext: %v", err)
		return fmt.Errorf("failed to create post: %w", err)
	}

	//log.Printf("Created post %s ", post.Title)
	return nil
}

// const PostRepo_createPostQuery = `INSERT INTO posts (title, content, author_id, status, publish_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
const PostRepo_createPostQuery = `
 INSERT INTO posts (
  title,
  content,
  author_id,
  status,
  publish_at,
  created_at,
  updated_at
 )
 VALUES ($1, $2, $3, $4, $5, $6, $7)
 RETURNING id
`

const PostRepo_getPostByIDQuery = `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
 FROM posts
 WHERE id = $1
`

// GetByID получает пост по ID
func (r *PostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	//		query := `
	//	       SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
	//	       FROM posts
	//	       WHERE id = $1
	//	   `
	//log.Printf("(r *PostRepo) GetByID SQL: %q", PostRepo_getPostByIDQuery)
	var post model.Post
	//
	err := r.db.QueryRowContext(ctx, PostRepo_getPostByIDQuery, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AuthorID,
		&post.Status,
		&post.PublishAt,
		&post.CreatedAt,
		&post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, blogerrors.ErrPostNotFound // Возвращаем ошибку, если пост не найден
		}
		return nil, fmt.Errorf("GetByID failed to get post: %w", err) // Обработка других ошибок
	}

	return &post, nil // Возвращаем найденный пост
}

const PostRepo_getallQuery = `SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
 FROM posts
 WHERE status = 'published'
 ORDER BY created_at DESC
 LIMIT $1 OFFSET $2
`

// GetAll получает все посты с пагинацией
func (r *PostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {

	// TODO: Выполнить запрос
	rows, err := r.db.QueryContext(ctx, PostRepo_getallQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	// Итерировать по результатам
	var posts []*model.Post
	for rows.Next() {
		var post model.Post

		// Сканируем данные в структуру поста
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Status, &post.PublishAt, &post.CreatedAt, &post.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		//log.Printf("\t\t post: %+v", post) //
		posts = append(posts, &post)
	}

	// Проверяем на ошибки после итерации
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during row iteration: %w", err)
	}

	return posts, nil
}

// GetTotalCount получает общее количество постов
func (r *PostRepo) GetTotalCount(ctx context.Context) (int, error) {
	// TODO: Реализовать подсчет общего количества постов
	// HINT: Используйте SELECT COUNT(*) FROM posts

	query := `SELECT COUNT(*) FROM posts WHERE status = 'published'`

	var count int
	// TODO: Выполнить запрос и получить количество
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count of posts: %w", err) // Обработка ошибки
	}

	return count, nil // Возвращаем общее количество постов
}

// Update обновляет пост
func (r *PostRepo) Update(ctx context.Context, post *model.Post) error {
	query := `
		UPDATE posts
		SET title = $1, content = $2, updated_at = $3
		WHERE id = $4
	`
	post.UpdatedAt = time.Now()
	result, err := r.db.ExecContext(ctx, query, post.Title, post.Content, post.UpdatedAt, post.ID)
	if err != nil {
		return fmt.Errorf("failed to update post: %w", err) // Обработка ошибки
	}

	// Проверяем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err) // Обработка ошибки
	}

	// Если ни одна строка не была затронута, возвращаем ошибку
	if rowsAffected == 0 {
		return blogerrors.ErrPostNotFound // Возвращаем ошибку, если пост не найден
	}

	return nil // Возвращаем nil, если обновление прошло успешно
}

// Delete удаляет пост
func (r *PostRepo) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM posts WHERE id = $1`

	// Выполняем запрос с помощью ExecContext
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err) // Обработка ошибки
	}

	// Проверяем количество затронутых строк
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err) // Обработка ошибки
	}

	// Если ни одна строка не была затронута, возвращаем ошибку
	if rowsAffected == 0 {
		return blogerrors.ErrPostNotFound // Возвращаем ошибку, если пост не найден
	}

	return nil // Возвращаем nil, если удаление прошло успешно
}

// Exists проверяет существование поста
func (r *PostRepo) Exists(ctx context.Context, id int) (bool, error) {

	query := `SELECT EXISTS(SELECT 1 FROM posts WHERE id = $1)`

	var exists bool
	// Выполняем запрос с помощью QueryRowContext
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check post existence: %w", err) // Обработка ошибки
	}

	return exists, nil // Возвращаем результат проверки
}

// GetTotalCount получает общее количество постов автора
func (r *PostRepo) CountByAuthorID(ctx context.Context, authorID int) (int, error) {
	query := `SELECT COUNT(*) FROM posts 
				WHERE author_id = $1 
				AND status = 'published'`

	var count int

	err := r.db.QueryRowContext(ctx, query, authorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total count of posts by AuthorID: %w", err) // Обработка ошибки
	}

	return count, nil // Возвращаем общее количество постов
}

// GetByAuthorID получает посты определенного автора
func (r *PostRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	// TODO: (Опционально) Реализовать получение постов автора
	// Аналогично GetAll, но с дополнительным условием WHERE author_id = $X

	// Prepare the SQL query to fetch posts by author ID with pagination
	query := `
        SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
        FROM posts
        WHERE author_id = $1 AND status = 'published'
        ORDER BY created_at DESC
        LIMIT $2 OFFSET $3
    `

	// Execute the query
	rows, err := r.db.QueryContext(ctx, query, authorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("error querying posts: %w", err)
	}
	defer rows.Close()

	// Slice to hold the retrieved posts
	var posts []*model.Post

	// Iterate over the result set
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AuthorID, &post.Status, &post.PublishAt, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("error scanning post: %w", err)
		}
		posts = append(posts, &post)
	}

	// Check for errors from iterating over rows
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return posts, nil //return nil, fmt.Errorf("not implemented")
}

// GetScheduledPosts получает посты, готовые к публикации
func (r *PostRepo) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {

	query := `
        SELECT id, title, content, author_id, status, publish_at, created_at, updated_at
        FROM posts
        WHERE status = 'draft'
		AND publish_at IS NOT NULL 
		AND publish_at <= NOW()
        ORDER BY publish_at ASC
    `

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post // Инициализируем как пустой срез
	for rows.Next() {
		var post model.Post
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AuthorID,
			&post.Status,
			&post.PublishAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		)

		if err != nil {
			//log.Printf("\t\tGetScheduledPosts( failed to scan post...")
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}

		//log.Printf("\t\tGetScheduledPosts( + post: %d", post.ID)
		posts = append(posts, &post)
	}

	if err = rows.Err(); err != nil {
		//log.Printf("\t\tGetScheduledPosts( failed to iterate posts")
		return nil, fmt.Errorf("failed to iterate posts: %w", err)
	}

	//log.Printf("\t\tGetScheduledPosts( return ")
	return posts, nil // Всегда возвращаем срез, даже если он пустой
}

// PublishPost публикует пост
func (r *PostRepo) PublishPost(ctx context.Context, id int) error {
	query := `
		UPDATE posts
		SET status = 'published', publish_at = NULL, updated_at = $1
		WHERE id = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to publish post: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return blogerrors.ErrPostNotFound
	}

	return nil
}

// GetTotalCountByAuthorID получает общее количество постов автора status = 'published'
func (r *PostRepo) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	query := `SELECT COUNT(*) FROM posts WHERE author_id = $1 AND status = 'published'`

	var count int
	err := r.db.QueryRowContext(ctx, query, authorID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get total post count by author: %w", err)
	}

	return count, nil
}
