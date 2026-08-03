package service

import (
	blogerrors "blog-api-dp/internal/erros"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"blog-api-dp/pkg/auth"
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"time"
)

type UserService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// возвращает ErrUserAlreadyExists при любой ошибке ExistsByEmail.
// проверка  в тесте TestUserService_Register_EmailCheckError.
func (s *UserService) Register(ctx context.Context, req *model.UserCreateRequest) (*model.TokenResponse, error) {
	// Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	// Проверка уникальности email
	exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
	if err != nil {
		return nil, blogerrors.ErrUserAlreadyExists // handler будет сравнивать с этим sentinel error, чтобы вернуть 409 Conflict
	}
	if exists {
		return nil, blogerrors.ErrUserAlreadyExists // handler будет сравнивать с этим sentinel error, чтобы вернуть 409 Conflict
	}

	// Проверка уникальности username
	exist, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("ошибка при проверке имени пользователя: %w", err)
	}
	if exist {
		return nil, blogerrors.ErrUserAlreadyExists // handler будет сравнивать с этим sentinel error, чтобы вернуть 409 Conflict
	}

	// Хеширование пароля
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("ошибка при хешировании пароля: %w", err)
	}

	//  Создание модели пользователя
	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Сохранение пользователя
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("ошибка при сохранении пользователя: %w", err)
	}

	// Генерация JWT токена
	token, _, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при генерации токена: %w", err)
	}

	return &model.TokenResponse{
		Token: token,
		User:  model.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email}}, nil
}

// Реализовать вход пользователя
// ВАЖНО: При ошибке не раскрывать, что именно неправильно (email или пароль)
// возвращает ошибки: ErrInvalidCredentials, ...
func (s *UserService) Login(ctx context.Context, req *model.UserLoginRequest) (*model.TokenResponse, error) {

	// 1. Валидация входных данных
	if err := req.Validate(); err != nil {
		return nil, err
	}
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}

	// Найти пользователя по email через репозиторий
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		// В случае ошибки базы или если пользователь не найден, скрываем детали
		// fix: В Login похожая ситуация: handler ждёт service.ErrInvalidCredentials, но сервис возвращает обычные текстовые ошибки.
		return nil, blogerrors.ErrInvalidCredentials // Возвращаем sentinel error, чтобы handler мог вернуть 401 Unauthorized
	}

	if user == nil { // Пользователь не найден
		return nil, blogerrors.ErrInvalidCredentials // Возвращаем sentinel error, чтобы handler мог вернуть 401 Unauthorized
	}

	// Проверить пароль используя функцию из пакета auth
	if ok := auth.CheckPassword(req.Password, user.Password); !ok { // Ошибка проверки пароля — скрываем детали
		return nil, blogerrors.ErrInvalidCredentials // Возвращаем sentinel error, чтобы handler мог вернуть 401 Unauthorized
	}

	// 4. Генерация JWT при успешной аутентификации
	token, _, err := s.jwtManager.GenerateToken(user.ID, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании токена") // Можно скрыть, если считается, что это внутреняя ошибка
	}

	// 5. Возвращаем TokenResponse
	return &model.TokenResponse{
		Token: token,
		User:  model.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email}}, nil
}

// Получить пользователя по ID через репозиторий
func (s *UserService) GetByID(ctx context.Context, id int) (*model.User, error) {

	user, err := s.userRepo.GetByID(ctx, id) // fix: Получаем пользователя через репозиторий
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("пользователь с ID %d не найден", id)
	}

	return user, nil
}

// Получить пользователя по e-mail
func (s *UserService) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email) // fix: Получаем пользователя через репозиторий
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении пользователя по email: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("пользователь с email %s не найден", email)
	}

	return user, nil
}

// ValidateEmail проверяет формат email (базовая проверка), возвращает ошибку если Email не соответствует требованиям
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("no e-mail provided")
	}

	const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	matched, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		return fmt.Errorf("error checking email format: %v", err)
	}
	if !matched {
		return fmt.Errorf("invalid e-mail format")
	}

	return nil
}
