package blogerrors

import (
	"encoding/json"
	"errors"
	"net/http"
)

// 409 при уже существующем пользователе,
// 401 при неверных данных входа,
// 403 при попытке редактировать чужой пост,
// 404 при отсутствии сущности.
var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	// user
	ErrUserAlreadyExists = errors.New("user already exists") // - ошибка, возникающая при попытке зарегистрировать пользователя, который уже существует.
	ErrUserNotFound      = errors.New("user not found")
	// post
	ErrInvalidPostID = errors.New("invalid post ID")
	ErrPostNotFound  = errors.New("post not found")
	ErrPostNotExists = errors.New("post does not exist")
	// comment
	ErrUnauthorized    = errors.New("unauthorized")
	ErrForbidden       = errors.New("forbidden")
	ErrCommentNotFound = errors.New("comment not found")
	ErrContextCanceled = errors.New("canceled")
)

// ErrorResponse - структура для ответа с ошибкой
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// ReplyError отправляет ошибку в JSON-ответе
func ReplyJsonError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	})

	// errorResponse := map[string]string{"error": fmt.Sprintf("(%d) %s", statusCode, message)}
	//	if err := json.NewEncoder(w).Encode(errorResponse); err != nil {
	//		log.Printf("Failed to write error response: %v", err)

}
