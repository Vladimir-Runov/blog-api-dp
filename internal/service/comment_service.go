package service

import (
	blogerrors "blog-api-dp/internal/erros"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
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
	log.Printf("CommentService: Creating comment for user ID %d on post ID %d", userID, req.PostID)
	// 1. Валидация данных
	if err := req.Validate(); err != nil {
		return nil, err
	}

	//  Проверить, что пост существует
	exists, err := s.postRepo.Exists(ctx, req.PostID)
	if err != nil {
		log.Printf("CommentService: Failed to check existence of post ID %d: %v", req.PostID, err)
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

	// Сохранение через репозиторий
	if err := s.commentRepo.Create(ctx, comment); err != nil {
		return nil, fmt.Errorf("failed to save comment: %w", err)
	}

	// 5. Опционально: обогащение ответа информацией об авторе
	author, err := s.userRepo.GetByID(ctx, userID)
	if err == nil {
		comment.AuthorID = author.ID
	}

	// 6. Вернуть созданный комментарий
	return comment, nil
}

// TODO: Получить комментарий по ID
func (s *CommentService) GetByID(ctx context.Context, id int) (*model.Comment, error) {
	// 1. Получить комментарий через репозиторий
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, blogerrors.ErrCommentNotFound // Return a specific error if the comment is not found
	}

	// Step 2: Check if the comment exists
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

const (
	defaultLimit = 20
	maxLimit     = 100
)

// Получить комментарии к посту с пагинацией
func (s *CommentService) GetByPost(ctx context.Context, postID int, limit, offset int) ([]*model.Comment, int, error) {
	// 1. Валидировать параметры пагинации (limit по умолчанию 20, максимум 100)
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
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
	log.Fatalf("CommentService.Update")
	// 1. Найти существующий комментарий
	commentUpd, err := s.commentRepo.GetByID(ctx, CommentId)
	if err != nil {
		return nil, err // Возвращаем ошибку, если комментарий не найден
	}

	log.Printf("Comment author id %d,  %d user id from context", commentUpd.AuthorID, userID)
	// 2. Проверить что userID является автором (иначе ErrForbidden)
	if commentUpd.AuthorID != userID {
		log.Printf("Comment author id %d <> %d user id from context", commentUpd.AuthorID, userID)
		return nil, blogerrors.ErrForbidden // fix: Возвращаем ErrForbidden 403 при попытке редактировать чужой пост,
	}

	// 3. Валидировать новый content
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// 4. Обновить content и временную метку
	commentUpd.Content = req.Content
	commentUpd.UpdatedAt = time.Now()

	//	updReq := model.CommentUpdateRequest{} // todo

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
	// TODO: Удалить комментарий
	// Шаги:
	// 1. Найти комментарий и проверить существование
	// 2. Проверить что userID является автором
	// 3. Удалить через репозиторий
	// 4. Вернуть соответствующую ошибку при неудаче
	// 1. Найти комментарий и проверить существование
	comment, err := s.commentRepo.GetByID(ctx, id)
	if err != nil {
		return err // Возвращаем ошибку, если комментарий не найден
	}

	// 2. Проверить, что userID является автором
	if comment.AuthorID != userID {
		return errors.New("user is not the author of the comment")
	}

	// 3. Удалить через репозиторий
	err = s.commentRepo.Delete(ctx, id)
	if err != nil {
		return err // Возвращаем ошибку, если удаление не удалось
	}

	// 4. Возвращаем nil, если удаление прошло успешно
	return nil //return fmt.Errorf("not implemented")
}

func (s *CommentService) GetByAuthor(ctx context.Context, authorID int, limit, offset int) ([]*model.Comment, int, error) {
	// TODO: Получить комментарии конкретного автора
	// Шаги:
	// 1. Валидировать параметры пагинации
	// 2. Получить комментарии автора через репозиторий
	// 3. Получить общее количество комментариев автора
	// 4. Опционально: добавить информацию об авторе
	// 5. Вернуть результат с общим количеством
	// 1. Валидировать параметры пагинации
	if limit <= 0 {
		return nil, 0, errors.New("limit must be greater than zero")
	}
	if offset < 0 {
		return nil, 0, errors.New("offset cannot be negative")
	}

	//	** todo
	//	// 2. Получить комментарии автора через репозиторий
	//	comments, err := s.commentRepo.GetCommentsByAuthor(ctx, authorID, limit, offset)
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//
	//	// 3. Получить общее количество комментариев автора
	//	totalCount, err := s.commentRepo.GetTotalCommentsByAuthor(ctx, authorID)
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//
	//	// 4. Опционально: добавить информацию об авторе (если необходимо)
	//
	//	// 5. Вернуть результат с общим количеством
	//	return comments, totalCount, nil //
	return nil, 0, fmt.Errorf("not implemented")
}

// validateCommentCreateRequest проверяет корректность данных для создания комментария
func validateCommentCreateRequest(req *model.CommentCreateRequest) error {
	// TODO: Реализовать валидацию content и PostID
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
	// TODO: Реализовать валидацию content
	// Проверка на пустое содержимое
	if strings.TrimSpace(req.Content) == "" {
		return errors.New("content cannot be empty")
	}

	// Проверка длины содержимого
	const maxContentLength = 500 // максимальная длина комментария
	if len(req.Content) > maxContentLength {
		return errors.New("content exceeds maximum length")
	}

	// Здесь можно добавить дополнительные проверки, если необходимо

	return nil
}
