package service

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

const (
	ConstDefaultLimit = 20
	ConstMaxLimit     = 100
)

type CommentService struct {
	commentRepo repository.CommentRepository
	postRepo    repository.PostRepository
	userRepo    repository.UserRepository
}

func NewCommentService(
	commentRepo repository.CommentRepository,
	postRepo repository.PostRepository,
	userRepo repository.UserRepository,
) *CommentService {
	return &CommentService{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		userRepo:    userRepo,
	}
}

// TODO: Создать новый комментарий
func (s *CommentService) Create(ctx context.Context, userID int, req *model.CommentCreateRequest) (*model.Comment, error) {

	if err := req.Validate(); err != nil {
		return nil, err
	}

	//  Проверить, что пост существует
	exists, err := s.postRepo.Exists(ctx, req.PostID)
	if err != nil {
		log.Printf("CommentService: Error postRepo.Exists post ID %d: %v", req.PostID, err)
		return nil, err // Возвращаем SQL ошибку, если не удалось проверить существование поста
	}

	if !exists {
		log.Printf("CommentService: Post not found for comment creation.  post ID %d", req.PostID)
		return nil, blogerrors.ErrPostNotFound // Возвращаем ошибку, если пост не найден
	}

	// Создать модель комментария с userID как автором
	comment := &model.Comment{
		AuthorID:  userID,
		PostID:    req.PostID,
		Content:   req.Content,
		CreatedAt: time.Now(), // добавляем время создания
	}

	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to save comment: %w", err)
	}

	author, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		comment.AuthorID = author.ID
	}

	return comment, nil
}

// TODO: Получить комментарий по ID
func (s *CommentService) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	log.Printf("\n\n\t(s *CommentService) GetByID(....%d", id)

	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, blogerrors.ErrCommentNotFound // Return a specific error if the comment is not found
	}
	if comment == nil {
		return nil, blogerrors.ErrCommentNotFound // Return a specific error if the comment is not found
	}

	// 2. Опционально: добавить информацию об авторе
	author, err := s.userRepo.GetByID(ctx, comment.AuthorID)
	if err == nil { // если нашлось, добавляем, иначе оставляем как есть (не ошибка)
		comment.AuthorID = author.ID
	}

	return comment, nil // 3. Вернуть результат или ErrCommentNotFound
}

// Получить комментарии к посту с пагинацией
func (s *CommentService) GetByPost(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, int, error) {
	// 1. Валидировать параметры пагинации (limit по умолчанию 20, максимум 100)
	if limit <= 0 {
		limit = ConstDefaultLimit
	}
	if limit > ConstMaxLimit {
		limit = ConstMaxLimit
	}

	// 2. Опционально: проверить существование поста
	exists, err := s.postRepo.Exists(ctx, postID)
	if err != nil {
		return nil, 0, err // Обработка ошибки СУБД
	}
	if !exists {
		return nil, 0, blogerrors.ErrPostNotFound
	}

	// 3. Получить комментарии через репозиторий
	comments, err := s.commentRepo.GetByPostID(ctx, postID, limit, offset)
	if err != nil {
		return nil, 0, err // Возвращаем ошибку, если не удалось получить комментарии
	}

	// 4. Получить общее количество для пагинации
	totalCount, err := s.commentRepo.GetCountByPostID(ctx, postID)
	if err != nil {
		return nil, 0, err // Возвращаем ошибку, если не удалось получить общее количество
	}

	// 5. Опционально: обогатить данные информацией об авторах
	// Например, можно добавить информацию о пользователях в комментарии

	// 6. Вернуть комментарии и общее количество
	return comments, totalCount, nil
}

// Обновить комментарий
func (s *CommentService) Update(ctx context.Context, CommentId int, userID int, req *model.CommentUpdateRequest) (*model.Comment, error) {
	log.Printf("\t(s *CommentService) Update(....CommentId=%d, userID=%d \n\n", CommentId, userID)
	commentUpd, err := s.commentRepo.GetByID(ctx, CommentId)
	if err != nil {
		return nil, blogerrors.ErrCommentNotFound // Return a specific error (CommentNotFound ) if the comment is not found or SQL
	}

	if commentUpd.AuthorID != userID { // userID является автором (иначе ErrForbidden)
		log.Printf("Comment author id %d <> %d user id from context\n\n", commentUpd.AuthorID, userID)
		return nil, blogerrors.ErrForbidden // fix: Возвращаем ErrForbidden 403 при попытке редактировать чужой коммент
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	commentUpd.Content = req.Content
	commentUpd.UpdatedAt = time.Now()

	// 5. Сохранить через репозиторий
	err = s.commentRepo.Update(ctx, commentUpd) // hot fix: вызов репозитория для обновления комментария (испр. рекурсивный вызов)
	if err != nil {
		return nil, err // Возвращаем ошибку, если обновление не удалось
	}

	// 6. Опционально: добавить информацию об авторе (если требуется)

	// 7. Вернуть обновленный комментарий
	return commentUpd, nil
}

func (s *CommentService) Delete(ctx context.Context, id int, userID int) error {
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return blogerrors.ErrCommentNotFound //  NotFound , если комментарий не найден
	}

	if comment.AuthorID != userID {
		return blogerrors.ErrForbidden //  ErrForbidden, если пользователь не является автором комментария
	}

	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return err // Возвращаем ошибку, если удаление не удалось
	}

	return nil
}

// validateCommentCreateRequest проверяет корректность данных для создания комментария
func validateCommentCreateRequest(req *model.CommentCreateRequest) error {

	// Проверка на пустое содержимое
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// Проверка длины содержимого
	const maxContentLength = 500 // максимальная длина комментария
	if len(req.Content) > maxContentLength {
		return errors.New("content exceeds maximum length")
	}

	// Проверка на наличие PostID
	if req.PostID <= 0 {
		return errors.New("invalid PostID")
	}

	return nil
}

// validateCommentUpdateRequest проверяет корректность данных для обновления комментария
func validateCommentUpdateRequest(req *model.CommentUpdateRequest) error {
	// Проверка на пустое содержимое
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// Проверка длины содержимого
	const maxContentLength = 500 // максимальная длина комментария
	if len(req.Content) > maxContentLength {
		return errors.New("content exceeds maximum length")
	}

	return nil
}
