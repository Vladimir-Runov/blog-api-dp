package handler

// https://github.com/Vladimir-Runov/blog-api-dp

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/service"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

type PostHandler struct {
	postService *service.PostService
}

func NewPostHandler(postService *service.PostService) *PostHandler {
	return &PostHandler{
		postService: postService,
	}
}

// Create - обрабатывает создание нового поста
// POST /api/posts
// Требует аутентификации
func (h *PostHandler) Create(w http.ResponseWriter, r *http.Request) {
	log.Printf("POST /api/posts: Creating post")
	if r.Method != http.MethodPost {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		blogerrors.ReplyJsonError(w, "Unauthorized (Create)", http.StatusUnauthorized)
		return
	}

	// 3. Декодирование JSON тела в структуру PostCreateRequest
	var req model.PostCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		blogerrors.ReplyJsonError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	post, err := h.postService.Create(r.Context(), userID, &req)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Failed to create post", http.StatusBadRequest) // 400 (было 500)
		return
	}

	// 5. Отправка созданного поста как JSON с статусом 201
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		blogerrors.ReplyJsonError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetByID - возвращает пост по ID
// GET /api/posts/{id}
// Не требует аутентификации
func (h *PostHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /api/posts/{id}: Retrieving post")
	// 1. Проверить метод запроса (должен быть GET)
	if r.Method != http.MethodGet {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Извлечь ID из URL
	// Примерный URL: /api/posts/123
	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest)
		return
	}

	// 3. Получить пост через postService.GetByID
	post, err := h.postService.GetByID(r.Context(), postID)
	if err != nil { // 4. Обработать ошибки (ErrPostNotFound -> 404)
		if err == blogerrors.ErrPostNotFound {
			blogerrors.ReplyJsonError(w, "Post not found", http.StatusNotFound) //  404 при отсутствии поста.
			return
		}
		blogerrors.ReplyJsonError(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Printf("Retrieving post with ID: %d", postID)

	// 5. Вернуть пост как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	if err := json.NewEncoder(w).Encode(post); err != nil {
		blogerrors.ReplyJsonError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// GetAll - возвращает список постов с пагинацией
// GET /api/posts?limit=10&offset=0
// Не требует аутентификации
func (h *PostHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	log.Printf("GET /api/posts: GetAll - Retrieving all posts with pagination")

	if r.Method != http.MethodGet {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	limit := service.ConstDefaultLimit // значение по умолчанию
	offset := 0

	// Парсим limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	// Парсим offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// 3. Получить посты через postService.GetAll
	posts, total, err := h.postService.GetAll(r.Context(), limit, offset)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Failed to fetch posts:", http.StatusInternalServerError)
		return
	}
	if posts == nil {
		log.Printf(" = nil")
		posts = []*model.Post{}
	}

	// 4. Создать ответ с метаданными пагинации
	response := struct {
		Items      []*model.Post `json:"items"`
		TotalCount int           `json:"total_count"`
		Limit      int           `json:"limit"`
		Offset     int           `json:"offset"`
	}{
		Items:      posts,
		TotalCount: total,
		Limit:      limit,
		Offset:     offset,
	}

	// 5. Вернуть список постов как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK

	if err := json.NewEncoder(w).Encode(response); err != nil {
		blogerrors.ReplyJsonError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Update - обновляет пост
// PUT /api/posts/{id}
// Требует аутентификации, может обновить только автор
func (h *PostHandler) Update(w http.ResponseWriter, r *http.Request) {
	log.Printf("PUT /api/posts/{id} Updating post")

	// 1. Проверить метод запроса (должен быть PUT)
	if r.Method != http.MethodPut {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Получаем userID из контекста (только авторы могут обновлять)
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		log.Printf("Unauthorized access attempt to update post")
		blogerrors.ReplyJsonError(w, "Unauthorized (Update)", http.StatusUnauthorized) // 401 Unauthorized
		return
	}

	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest) //
		return
	}

	// 4. Декодируем JSON тело в PostUpdateRequest
	var req model.PostUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: {%s} ( %v )", r.Body, err)
		blogerrors.ReplyJsonError(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// 5. Обновляем через postService.Update
	updatedPost, err := h.postService.Update(r.Context(), postID, userID, &req)
	if err != nil { // Обработать ошибки (404 для не найден, 403 для чужого поста)

		if errors.Is(err, blogerrors.ErrPostNotFound) {
			blogerrors.ReplyJsonError(w, "Post Not Found", http.StatusNotFound) // 404 при отсутствии поста.
			return
		}

		if errors.Is(err, blogerrors.ErrForbidden) {
			blogerrors.ReplyJsonError(w, "Forbidden", http.StatusForbidden) // 403 при попытке обновить чужой пост.
			return
		}

		blogerrors.ReplyJsonError(w, "Failed to Update ", http.StatusBadRequest) // 400 Bad Request
		return
	}

	// 6. Возвращаем обновленный пост как JSON (200 OK)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(updatedPost)
}

// Delete - удаляет пост
// DELETE /api/posts/{id}
// Требует аутентификации, может удалить только автор
func (h *PostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.Printf("DELETE /api/posts/{id}: Deleting post")

	// 1. Проверить метод запроса (должен быть DELETE)
	if r.Method != http.MethodDelete {
		blogerrors.ReplyJsonError(w, "Method Not Allowed", http.StatusMethodNotAllowed) // 405 Method Not Allowed
		return
	}

	// 2. Получение userID из контекста
	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		blogerrors.ReplyJsonError(w, "Unauthorized (Delete)", http.StatusUnauthorized) // 401 Unauthorized
		return
	}

	// 3. Извлечение ID поста из URL
	// Примерный URL: /api/posts/123
	idStr := extractIDFromPath(r.URL.Path, "/api/posts/")
	postID, err := strconv.Atoi(idStr)
	if err != nil {
		log.Printf("Invalid post ID: %v", err)
		blogerrors.ReplyJsonError(w, "Invalid post ID", http.StatusBadRequest) // 400 Bad Request
		return
	}

	// 4. Удаление поста через сервис
	err = h.postService.Delete(r.Context(), postID, userID)
	if err != nil { // 5. Обработать ошибки (404 для не найден, 403 для чужого поста)
		if err == blogerrors.ErrPostNotFound {
			log.Printf("*** Post not found: %v", err)
			blogerrors.ReplyJsonError(w, "Post not found!", http.StatusNotFound) //  404 при отсутствии поста.
			return
		}
		if err == blogerrors.ErrForbidden {
			log.Printf("Forbidden access attempt to delete post")
			blogerrors.ReplyJsonError(w, "Forbidden", http.StatusForbidden) //  403 при попытке удалить чужой пост.
			return
		}
		log.Printf("Internal server error: %v", err)
		blogerrors.ReplyJsonError(w, "Internal Server Error", http.StatusInternalServerError) // 500 Internal Server Error
		return
	}

	// Вернуть 204 No Content при успехе
	w.WriteHeader(http.StatusNoContent) // 204 No Content
}

// GetByAuthor - возвращает посты конкретного автора
// GET /api/posts/author/{authorID}?limit=10&offset=0
// Не требует аутентификации
func (h *PostHandler) GetByAuthor(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authorIDStr := chi.URLParam(r, "authorID")
	authorID, err := strconv.Atoi(authorIDStr)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	limit := 10
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	posts, total, err := h.postService.GetByAuthor(r.Context(), authorID, limit, offset)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Invalid author ID", http.StatusBadRequest)
		return
	}

	type PostsResponse struct {
		Posts    []*model.Post `json:"posts"`
		Total    int           `json:"total"`
		Limit    int           `json:"limit"`
		Offset   int           `json:"offset"`
		AuthorID int           `json:"author_id"`
	}

	resp := PostsResponse{
		Posts:    posts,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
		AuthorID: authorID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

}

// extractIDFromPath извлекает ID из пути URL
func extractIDFromPath(path, prefix string) string {
	// TODO: Реализовать извлечение ID из пути URL
	// Пример: path = "/api/posts/123", prefix = "/api/posts/" Должен вернуть "123"

	// Убираем префикс из начала пути
	if len(path) <= len(prefix) || path[:len(prefix)] != prefix {
		return ""
	}
	idPart := path[len(prefix):]
	idPart = strings.Trim(idPart, "/")
	return idPart
}
