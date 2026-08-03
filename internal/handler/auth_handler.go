package handler

import (
	blogerrors "blog-api-dp/internal/erros"
	"blog-api-dp/internal/middleware"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/service"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type AuthHandler struct {
	userService *service.UserService
}

func NewAuthHandlerEx(userService *service.UserService) *AuthHandler {
	return &AuthHandler{
		userService: userService,
	}
}

// Register обрабатывает запрос на регистрацию нового пользователя
// POST /api/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		log.Printf("Register.Error invalid method: %s", r.Method)
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 2. Декодировать JSON тело в UserCreateRequest
	var req model.UserCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Register.Error decoding register request: %v", err)
		blogerrors.ReplyJsonError(w, "Bad request", http.StatusBadRequest)
		return
	}

	// 3. Вызвать userService.Register
	ctx := r.Context() // fix: должен использоваться r.Context().
	tokenResp, err := h.userService.Register(ctx, &model.UserCreateRequest{Username: req.Username, Password: req.Password, Email: req.Email})
	if err != nil {
		// fix: AuthHandler.Register сравнивает ошибку с model.ErrUserAlreadyExists, а UserService.Register при занятом email возвращает обычную ошибку через fmt.Errorf(“email уже занят”), а не этот sentinel error.
		// Поэтому handler не сможет распознать конфликт и вместо 409 Conflict вернёт внутреннюю ошибку.
		if err == blogerrors.ErrUserAlreadyExists { // 4. Обработать ошибки (ErrUserAlreadyExists -> 409 Conflict)
			log.Printf("Register.Error user already exists: %v", err)
			blogerrors.ReplyJsonError(w, "User already exists", http.StatusConflict) // 409 Conflict при попытке зарегистрировать уже существующего пользователя
			return
		}
		log.Printf("Register.Error registering user: %v", err)
		blogerrors.ReplyJsonError(w, "Internal server error (7)", http.StatusInternalServerError)
		return
	}

	// 5. Вернуть JSON ответ с токеном (201 Created)
	log.Printf("User registered successfully. Email: %s", req.Email)

	w.Header().Set("Content-Type", "application/json") // Заголовок  до отправки статуса
	w.WriteHeader(http.StatusCreated)                  // 201 Created при успешной регистрации
	json.NewEncoder(w).Encode(map[string]string{"token": tokenResp.Token})
}

// Login обрабатывает запрос на вход пользователя
// вход POST /api/login
// возвращает JSON с токеном при успешной аутентификации или ошибки: 405, 400, 401, 500
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {

	// 1. Проверить метод запроса (должен быть POST)
	if r.Method != http.MethodPost {
		blogerrors.ReplyJsonError(w, "Method not allowed", http.StatusMethodNotAllowed) // 405 Method Not Allowed при неверном методе запроса
		return
	}

	// 2. Декодировать JSON тело в UserLoginRequest
	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding login request: %v", err)
		blogerrors.ReplyJsonError(w, "Bad request", http.StatusBadRequest) // 400 Bad Request при некорректном JSON
		return
	}

	if err := req.Validate(); err != nil {
		log.Printf("Error validating login request: %v", err)
		blogerrors.ReplyJsonError(w, err.Error(), http.StatusBadRequest) // 400 Bad Request при некорректных данных
		return
	}

	// 3. Вызвать userService.Login
	ctx := r.Context() // fix: должен использоваться r.Context().
	tokenResp, err := h.userService.Login(ctx, &req)
	if err != nil { // 4. Обработать ошибки (ErrInvalidCredentials -> 401 Unauthorized)
		if err == blogerrors.ErrInvalidCredentials {
			log.Printf("Invalid credentials for user: %s", req.Email)
			blogerrors.ReplyJsonError(w, "Unauthorized", http.StatusUnauthorized) // 401 Unauthorized при неверных учетных данных
			return
		}
		// Отправляем клиенту сообщение об ошибке с описанием
		blogerrors.ReplyJsonError(w, fmt.Sprintf("Internal server error: %v", err.Error()), http.StatusInternalServerError) // 500 Internal Server Error при других ошибках
		return
	}

	// 5. Вернуть JSON ответ с токеном (200 OK)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"token": tokenResp.Token})
}

// GetProfile возвращает профиль текущего пользователя (опционально)
// Этот метод не используется в эталонной реализации
func (h *AuthHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// TODO: Опционально - реализовать получение профиля
	// Этот эндпоинт не обязателен для базовой реализации
	// http.Error(w, "Not implemented", http.StatusNotImplemented)

	userID, ok := getUserIDFromContext(r.Context())
	if !ok {
		blogerrors.ReplyJsonError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetByID(r.Context(), userID)
	if err != nil {
		blogerrors.ReplyJsonError(w, "Failed to fetch user profile", http.StatusInternalServerError)
		return
	}

	if user == nil {
		blogerrors.ReplyJsonError(w, "User not found", http.StatusNotFound)
		return
	}

	// Подготовка ответа - исключаем пароль и другие чувствительные данные
	response := &model.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	}

	// Отправляем ответ в формате JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		blogerrors.ReplyJsonError(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// getUserIDFromContext извлекает userID пользователя из контекста
// Ключ устанавливается в auth middleware
func getUserIDFromContext(ctx context.Context) (int, bool) {
	userID, ok := ctx.Value(middleware.UserIDKey).(int)
	return userID, ok
}
