package middleware

import (
	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/pkg/auth"
	"context"
	"encoding/json"

	"net/http"
	"strings"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// UserIDKey is the key for storing user ID in context
	UserIDKey contextKey = "userID"
	// UserEmailKey is the key for storing user email in context
	UserEmailKey contextKey = "userEmail"
	// UserNameKey is the key for storing username in context
	UserNameKey contextKey = "username"
)

// AuthMiddleware provides JWT authentication
type AuthMiddleware struct {
	jwtManager *auth.JWTManager
}

// NewAuthMiddleware creates a new auth middleware instance
func NewAuthMiddleware(jwtManager *auth.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{
		jwtManager: jwtManager,
	}
}

// RequireAuth is a middleware that requires valid JWT token
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Извлечь токен из заголовка Authorization (Bearer токен)
		tokenString := extractToken(r)
		if tokenString == "" {
			blogerrors.ReplyJsonError(w, "Unauthorized...missing token", http.StatusUnauthorized) //
			return
		}
		// 2. Валидировать токен через jwtManager
		claims, err := m.jwtManager.ValidateToken(tokenString)
		if err != nil { // 3. Обработать ошибки валидации
			blogerrors.ReplyJsonError(w, "Unauthorized...Invalid token ("+err.Error()+")", http.StatusUnauthorized) //
			return
		}

		// 4. Добавить данные пользователя в контекст
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
		ctx = context.WithValue(ctx, UserNameKey, claims.Username)
		r = r.WithContext(ctx)

		// 5. Передать управление следующему handler
		next.ServeHTTP(w, r)
	}
}

// не пойму, зачем эта функция нужна, если есть RequireAuth... где вызывать OptionalAuth?

// TODO: Реализовать опциональную проверку JWT токена
// OptionalAuth is a middleware that extracts JWT token if present, but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		tokenString := extractToken(r)
		if tokenString != "" {
			// 2. Валидировать токен через jwtManager
			claims, err := m.jwtManager.ValidateToken(tokenString)
			if err == nil {
				// 3. Если токен валидный - добавить данные в контекст
				ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
				ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
				ctx = context.WithValue(ctx, UserNameKey, claims.Username)
				r = r.WithContext(ctx)
			} else {
				// Обработка ошибок валидации (необязательно)
				// Можно логировать ошибку или игнорировать
			}

		}

		// 4. В любом случае передать управление следующему handler
		next.ServeHTTP(w, r) // Временная реализация next(w, r)
	}
}

// extractToken извлекает JWT токен из заголовка Authorization
func extractToken(r *http.Request) string {
	// TODO: Извлечь JWT токен из заголовка Authorization
	// Формат: "Bearer <token>"

	// Получаем значение заголовка Authorization  Попытаться извлечь токен из заголовка
	authHeader := r.Header.Get("Authorization")
	// Проверяем, что заголовок не пустой
	if authHeader != "" {
		// Разбиваем строку на части  Извлекаем Bearer токен)
		parts := strings.Split(authHeader, " ")

		if len(parts) == 2 && parts[0] == "Bearer" { // Проверяем, что формат правильный (Bearer <token>)
			return parts[1] // Возвращаем токен
		}
	}

	return "" // Возвращаем пустую строку, если токен не найден
}

// GetUserIDFromContext извлекает ID пользователя из контекста
// Извлечь userID из контекста (ключ UserIDKey)
func GetUserIDFromContext(ctx context.Context) (int, bool) {

	userID, ok := ctx.Value(UserIDKey).(int) // Извлекаем userID из контекста
	return userID, ok                        // Возвращаем ID и статус наличия
}

// GetUserEmailFromContext извлекает email пользователя из контекста
// Извлечь email из контекста (ключ UserEmailKey)
func GetUserEmailFromContext(ctx context.Context) (string, bool) {

	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

// GetUsernameFromContext извлекает username из контекста
// Извлечь username из контекста (ключ UserNameKey)
func GetUsernameFromContext(ctx context.Context) (string, bool) {

	username, ok := ctx.Value(UserNameKey).(string)
	return username, ok
}

// ErrorResponse представляет собой структуру для JSON ответа об ошибке
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// writeJSONError отправляет ошибку в формате JSON
func writeJSONError(w http.ResponseWriter, message string, statusCode int) {

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	// Создаем объект ErrorResponse
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}

	// Сериализуем объект в JSON и отправляем его
	if err := json.NewEncoder(w).Encode(response); err != nil {
		// Если произошла ошибка при кодировании, отправляем стандартную ошибку
		//http.Error(w, "Internal Server Error (22)", http.StatusInternalServerError)
		blogerrors.ReplyJsonError(w, "Internal server error.", http.StatusInternalServerError) //
	}
}

// Вспомогательные функции для упрощения использования middleware

// Chain позволяет объединить несколько middleware в цепочку
func Chain(handler http.HandlerFunc, middlewares ...func(http.HandlerFunc) http.HandlerFunc) http.HandlerFunc {
	// TODO: Реализовать объединение middleware в цепочку
	// Применить их в правильном порядке
	// ... middleware в обратном порядке
	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}
	return handler
}
