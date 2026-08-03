package service

import (
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"blog-api-dp/pkg/auth"
	"context"
	"fmt"
	"time"
)

type PostService struct {
	postRepo   repository.PostRepository
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewPostService(postRepo repository.PostRepository, userRepo repository.UserRepository, jwtManager *auth.JWTManager) *PostService {
	return &PostService{
		postRepo:   postRepo,
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// TODO: Создать новый пост
func (s *PostService) Create(ctx context.Context, userID int, req *model.PostCreateRequest) (*model.Post, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Определяем статус поста
	status := model.PostStatusPublished
	var publishAt *time.Time = nil

	if req.PublishAt != nil {
		if req.PublishAt.After(time.Now()) {
			status = model.PostStatusDraft
			publishAt = req.PublishAt
		} else {
			status = model.PostStatusPublished
		}
	}

	post := &model.Post{
		Title:     req.Title,
		Content:   req.Content,
		AuthorID:  userID,
		Status:    status,
		PublishAt: publishAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := s.postRepo.Create(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to save post: %w", err)
	}

	return post, nil
}

func (s *PostService) GetByID(ctx context.Context, id int) (*model.Post, error) {

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}

	if post.AuthorID != 0 {
		author, err := s.userRepo.GetByID(ctx, post.AuthorID)
		if err == nil {
			post.AuthorID = author.ID
		}
	}

	return post, nil
	//return nil, fmt.Errorf("not implemented")
}

func (s *PostService) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, int, error) {

	if limit <= 0 {
		limit = 10 // default limit
	} else if limit > 100 {
		limit = 100 // maximum limit
	}

	if offset < 0 {
		offset = 0 // ensure offset is not negative
	}

	// Step 2: Get posts from the repository
	posts, err := s.postRepo.GetAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts: %w", err)
	}

	// Step 3: Get total count for pagination
	totalCount, err := s.postRepo.GetTotalCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get total post count: %w", err)
	}

	for _, post := range posts {
		if post.AuthorID != 0 {
			author, err := s.userRepo.GetByID(ctx, post.AuthorID)
			if err == nil {
				post.AuthorID = author.ID
			}
		}
	}

	return posts, totalCount, nil
	//return nil, 0, fmt.Errorf("not implemented")
}

func (s *PostService) Update(ctx context.Context, id int, userID int, req *model.PostUpdateRequest) (*model.Post, error) {

	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return nil, fmt.Errorf("post not found")
	}

	// Step 2: Check that userID is the author
	if post.AuthorID != userID {
		return nil, fmt.Errorf("forbidden: you are not the author of this post")
	}

	// Step 3: Validate new data (if provided)
	if err := validatePostUpdateRequest(req); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Step 4: Update only changed fields
	if req.Title != "" {
		post.Title = req.Title
	}
	if req.Content != "" {
		post.Content = req.Content
	}

	err = s.postRepo.Update(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}

	return post, nil
	//return nil, fmt.Errorf("not implemented")
}

func (s *PostService) Delete(ctx context.Context, id int, userID int) error {
	// TODO: Удалить пост
	// Шаги:
	// 1. Найти пост и проверить существование
	// 2. Проверить что userID является автором
	// 3. Удалить через репозиторий
	// 4. Вернуть соответствующую ошибку при неудаче

	// Step 1: Find post and check existence
	post, err := s.postRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get post: %w", err)
	}
	if post == nil {
		return fmt.Errorf("post not found")
	}

	// Step 2: Check that userID is the author
	if post.AuthorID != userID {
		return fmt.Errorf("forbidden: you are not the author of this post")
	}

	// Step 3: Delete through repository
	if err := s.postRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	// Step 4: Return nil if successful
	return nil
	//return fmt.Errorf("not implemented")
}

func (s *PostService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, int, error) {

	if limit <= 0 {
		return nil, 0, fmt.Errorf("invalid limit: must be greater than 0")
	}
	if offset < 0 {
		return nil, 0, fmt.Errorf("invalid offset: must be non-negative")
	}

	posts, err := s.postRepo.GetByAuthorID(ctx, authorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get posts by author: %w", err)
	}

	totalPosts, err := s.postRepo.CountByAuthorID(ctx, authorID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts by author: %w", err)
	}

	return posts, totalPosts, nil
}

// GetScheduledPosts получает отложенные посты (для админов)
func (s *PostService) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
	return s.postRepo.GetScheduledPosts(ctx)
}

// validatePostCreateRequest проверяет корректность данных для создания поста
func validatePostCreateRequest(req *model.PostCreateRequest) error {
	// TODO: Реализовать валидацию title и content

	if req.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	// Check if content is empty
	if req.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	const maxTitleLength = 100
	const maxContentLength = 5000

	if len(req.Title) > maxTitleLength {
		return fmt.Errorf("title cannot exceed %d characters", maxTitleLength)
	}

	if len(req.Content) > maxContentLength {
		return fmt.Errorf("content cannot exceed %d characters", maxContentLength)
	}

	return nil
}

// validatePostUpdateRequest проверяет корректность данных для обновления поста
func validatePostUpdateRequest(req *model.PostUpdateRequest) error {
	// TODO: Реализовать валидацию опциональных полей
	// Validate Title if provided

	if req.Title == "" {
		return fmt.Errorf("title cannot be empty")
	}

	const maxTitleLength = 100
	if len(req.Title) > maxTitleLength {
		return fmt.Errorf("title cannot exceed %d characters", maxTitleLength)
	}

	// Validate Content if provided

	if req.Content == "" {
		return fmt.Errorf("content cannot be empty")
	}

	const maxContentLength = 5000
	if len(req.Content) > maxContentLength {
		return fmt.Errorf("content cannot exceed %d characters", maxContentLength)
	}

	return nil
}
