package auth

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

// Claims представляет данные, хранимые в JWT токене
type Claims struct {
	UserID               int    `json:"user_id"`
	Email                string `json:"email"`
	Username             string `json:"username"`
	jwt.RegisteredClaims        // TODO: Добавить стандартные JWT claims Подсказка: используйте jwt.RegisteredClaims или jwt.StandardClaims
}

// JWTManager управляет созданием и валидацией JWT токенов
type JWTManager struct {
	secretKey []byte
	ttl       time.Duration
}

// NewJWTManager создает новый экземпляр JWT менеджера
func NewJWTManager(secretKey string, ttlHours int) *JWTManager {
	// TODO: Инициализировать JWTManager
	// - Преобразовать secretKey в []byte
	// - Преобразовать ttlHours в time.Duration
	return &JWTManager{
		secretKey: []byte(secretKey),
		ttl:       time.Duration(ttlHours) * time.Hour,
	}
}

// GenerateToken создает новый JWT токен для пользователя
func (m *JWTManager) GenerateToken(userID int, email string, username string) (string, time.Time, error) {

	expirationTime := time.Now().Add(m.ttl) // Установить время истечения токена (текущее время + ttl)
	claims := Claims{
		UserID:   userID,
		Email:    email,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), //  время истечения
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // время выпуска
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims) //  Создать токен используя алгоритм подписи (например, HS256)
	signedToken, err := token.SignedString(m.secretKey)        // Подписать токен секретным ключом
	if err != nil {
		return "", time.Time{}, err // Вернуть ошибку в случае сбоя
	}

	// Вернуть подписанную строку токена и время истечения
	return signedToken, expirationTime, nil
}

// ValidateToken проверяет и парсит JWT токен
// ошибки:
// - Невалидная подпись -> ErrInvalidToken
// - Другие ошибки -> ErrInvalidToken
// - Истекший токен -> ErrExpiredToken
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {

	// Шаг 1: Распарсить токен с проверкой подписи
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secretKey, nil // Возвращаем секретный ключ для проверки подписи
	})

	if err != nil { // Обработка ошибок парсинга токена
		if err == jwt.ErrSignatureInvalid {
			return nil, ErrInvalidToken // Невалидная подпись
		}
		return nil, ErrInvalidToken // Другие ошибки
	}

	// Шаг 2: Извлечь claims из токена
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken // Если токен не валиден или не удалось извлечь claims
	}

	// Шаг 3: Проверить время истечения токена
	if claims.ExpiresAt.Time.Before(time.Now()) {
		log.Printf(" Token expired at %v", claims.ExpiresAt.Time)
		return nil, ErrExpiredToken // Истекший токен
	}

	// Шаг 4: Вернуть claims если токен валидный
	log.Printf("\tToken valid for userID: %d, e-mail: %s, username: %s", claims.UserID, claims.Email, claims.Username)
	return claims, nil
}

// RefreshToken обновляет существующий токен (опциональное задание)
// return "", time.Time{}, errors.New("not implemented")
func (m *JWTManager) RefreshToken(tokenString string) (string, time.Time, error) {
	// Шаг 1: Валидировать существующий токен
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return "", time.Time{}, err // Вернуть ошибку, если токен не валиден
	}

	// Шаг 2: Извлечь данные пользователя из старого токена
	userID := claims.UserID // Предполагается, что в claims есть поле UserID

	// Шаг 3: Сгенерировать новый токен с теми же данными
	expirationTime := time.Now().Add(m.ttl) // Установить время истечения токена (текущее время + ttl)
	newClaims := Claims{
		UserID:   userID,
		Email:    claims.Email,
		Username: claims.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime), //  время истечения
			IssuedAt:  jwt.NewNumericDate(time.Now()),     // время выпуска
		},
	}

	// Шаг 3: Создать токен используя алгоритм подписи (например, HS256)
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, newClaims)
	tokenString, err = newToken.SignedString(m.secretKey)
	if err != nil {
		return "", time.Time{}, err // Вернуть ошибку, если не удалось подписать новый токен
	}

	// Шаг 4: Вернуть новый токен и время истечения
	return tokenString, newClaims.ExpiresAt.Time, nil
}

// GetUserIDFromToken быстро извлекает ID пользователя из токена без полной валидации
func (m *JWTManager) GetUserIDFromToken(tokenString string) (int, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Проверка метода подписи
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return m.secretKey, nil
	})

	// Шаг 2: Проверить наличие ошибок при разборе токена
	if err != nil {
		return 0, err // Вернуть ошибку, если токен не может быть разобран
	}

	// Шаг 3: Извлечь UserID из токена
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userID, ok := claims["user_id"].(float64); ok { // Предполагается, что UserID хранится как float64
			return int(userID), nil // Вернуть UserID как int
		}
	}

	return 0, errors.New("user ID not found in token")
}
