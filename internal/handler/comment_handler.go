package handler

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/service"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{
		commentService: commentService,
	}
}

// Create - создание нового комментария
// POST /api/posts/{postId}/comments
// Требует аутентификации
func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST /api/comments: Creating comment...")
	if r.Method != http.MethodPost {
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		log.Printf("POST /api/comments: Unauthorized")
		blogerrors.ReplyJsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req model.CommentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("POST /api/comments: Invalid request body: %v", r.Body)
		blogerrors.ReplyJsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	idStr, err1 := extractIDFromCommentsPath(r.URL.Path, "/api/posts/") // Примерный URL: /api/comments/123
	if err1 != nil {
		log.Printf("POST /api/comments: Invalid post ID %s in URL: %s : %v", idStr, r.URL.Path, err1)
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	req.PostID, err = strconv.Atoi(idStr)
	if err != nil {
		log.Printf("POST /api/comments: Invalid post ID %d in URL: %s", req.PostID, r.URL.Path)
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	comment, err := h.commentService.Create(r.Context(), userID, &req)
	if err != nil {
		if errors.Is(err, blogerrors.ErrPostNotFound) {
			log.Printf("POST /api/comments: Post not found")
			blogerrors.ReplyJsonError(w, "Post not found", http.StatusNotFound) //  404 при отсутствии сущности.
		} else {
			log.Printf("POST /api/comments: Failed to create comment to post ID %d: %v", req.PostID, err)
			blogerrors.ReplyJsonError(w, "Failed to create comment", http.StatusBadRequest) // 400?   (было 500 Internal Server Error)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(comment)

	log.Printf("POST /api/comments: Comment created successfully")
}

// GetByID возвращает комментарий по ID
// GET /api/comments/{id}
// Не требует аутентификации
func (h *CommentHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /api/comments/{id}: Retrieving comment...")
	if r.Method != http.MethodGet {
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID из URL. Примерный URL: /api/comments/123
	idStr := extractIDFromPath(r.URL.Path, "/api/comments/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	// 3. Получить комментарий через сервис
	comment, err := h.commentService.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, blogerrors.ErrCommentNotFound) {
			blogerrors.ReplyJsonError(w, "Comment not found", http.StatusNotFound)
		} else {
			blogerrors.ReplyJsonError(w, "Failed to get comment", http.StatusInternalServerError) // 400 ?
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comment)
}

// GetByPost возвращает комментарии к посту
// GET /api/posts/{id}/comments?limit=20&offset=0
// curl -X GET "http://localhost:8080/api/posts/1/comments" -H "Authorization: Bearer " -H "Content-Type: application/json"
// Не требует аутентификации
func (h *CommentHandler) GetByPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet { // (должен быть GET)
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr, err1 := extractIDFromCommentsPath(r.URL.Path, "/api/posts/") // Примерный URL: /api/comments/123
	if err1 != nil {
		log.Printf("POST /api/comments: Invalid post ID %s in URL: %s : %v", idStr, r.URL.Path, err1)
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	postID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("POST /api/comments: Invalid post ID %d in URL: %s", postID, r.URL.Path)
		blogerrors.ReplyJsonError(w, "Invalid type (post ID)", http.StatusBadRequest)
		return
	}

	// Извлечь параметры пагинации
	query := r.URL.Query()
	limit, err := strconv.Atoi(query.Get("limit"))
	if err != nil || limit <= 0 {
		limit = service.ConstDefaultLimit // значение по умолчанию
	}

	offset, err := strconv.Atoi(query.Get("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	// Получить комментарии через сервис
	comments, totalCount, err := h.commentService.GetByPost(r.Context(), postID, limit, offset)
	if err != nil {
		if errors.Is(err, blogerrors.ErrPostNotFound) {
			blogerrors.ReplyJsonError(w, "Post not found", http.StatusNotFound)
		} else {
			blogerrors.ReplyJsonError(w, "Failed to get comments", http.StatusInternalServerError) // 400 ?
		}
		return
	}

	// 5. Создать ответ с метаданными
	resp := make([]model.CommentResponse, len(comments))
	for i, comment := range comments {
		resp[i] = model.CommentResponse{
			ID:      comment.ID,
			Content: comment.Content,
			PostID:  postID,
			//	Author:    ,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		}
	}

	// Отправить ответ
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	response := map[string]interface{}{
		"total":    totalCount,
		"comments": resp,
	}
	json.NewEncoder(w).Encode(response)
}

// Update обновляет комментарий
// PUT /api/comments/{id}
// Требует аутентификации, может обновить только автор
func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	log.Printf("PUT /api/comments/{id}: Edit comment...")
	if r.Method != http.MethodPut {
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		blogerrors.ReplyJsonError(w, "Unauthorized (comment.Update)", http.StatusUnauthorized)
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/comments/")
	commentID, err := strconv.Atoi(idStr)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	log.Printf("\tUser Id: %d updating comment %d", userID, commentID)

	// 4. Декодировать тело запроса
	var req model.CommentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		blogerrors.ReplyJsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	req.CommentID = commentID
	log.Printf("\tcall srv Update \n\n")

	// 5. Обновить комментарий через сервис
	comment, err := h.commentService.Update(r.Context(), commentID, userID, &req)
	if err != nil {
		if errors.Is(err, blogerrors.ErrCommentNotFound) {
			blogerrors.ReplyJsonError(w, "Comment not found", http.StatusNotFound) //
			return
		}
		if errors.Is(err, blogerrors.ErrForbidden) {
			blogerrors.ReplyJsonError(w, "You can only update your own comments", http.StatusForbidden) //
			return
		}

		blogerrors.ReplyJsonError(w, "Failed to update comment", http.StatusBadRequest) // 400, было 500
		return
	}

	// 6. Отправить обновленный комментарий
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.Printf("DELETE /api/comments/{id}: Deleting post")

	if r.Method != http.MethodDelete {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed) // 405 Method Not Allowed
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		blogerrors.ReplyJsonError(w, "Unauthorized (Delete)", http.StatusUnauthorized) // 401 Unauthorized
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/comments/")
	CommentID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid comment ID: %v", err)
		blogerrors.ReplyJsonError(w, "Invalid comment ID", http.StatusBadRequest) // 400 Bad Request
		return
	}

	err = h.commentService.Delete(r.Context(), CommentID, userID)
	if err != nil { // 5. Обработать ошибки (404 для не найден, 403 для чужого )
		if errors.Is(err, blogerrors.ErrCommentNotFound) {
			log.Printf("Comment not found: %v", err)
			blogerrors.ReplyJsonError(w, "Comment not found", http.StatusNotFound) //  404 при отсутствии .
			return
		}

		if errors.Is(err, blogerrors.ErrForbidden) {
			log.Printf("Forbidden access attempt to delete Comment")
			blogerrors.ReplyJsonError(w, "Forbidden", http.StatusForbidden) //  403 при попытке удалить чужой.
			return
		}

		log.Printf("Internal server error: %v", err.Error())
		blogerrors.ReplyJsonError(w, "Internal Server Error", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	// Вернуть 204 No Content при успехе
	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// extractIDFromPath извлекает ID из указанного пути, удаляя заданный префикс.
func extractIDFromCommentsPath(uri string, prefix string) (string, error) {
	// Убедимся, что путь начинается с заданного префикса
	if !strings.HasPrefix(uri, prefix) {
		return "", fmt.Errorf("path does not start with expected prefix: %s", prefix)
	}

	// Удаляем префикс из пути
	trimmedPath := strings.TrimPrefix(uri, prefix)

	// Извлекаем ID, который должен быть первым сегментом после префикса
	segments := strings.Split(trimmedPath, "/")
	if len(segments) > 0 {
		return segments[0], nil // Возвращаем первый сегмент как ID
	}

	return "", fmt.Errorf("no ID found in path after prefix: %s", prefix)
}
