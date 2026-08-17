package service

// go test ./internal/service
// go test ./internal/service -cover
import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	blogerrors "blog-api-dp/internal/errors"
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"blog-api-dp/pkg/auth"
)

type userRepoMock struct {
	// Встраивание интерфейса позволяет не реализовывать методы,
	// которые не используются в конкретном тесте.
	repository.UserRepository

	existsByEmailResult    bool
	existsByEmailErr       error
	existsByUsernameResult bool
	existsByUsernameErr    error

	getByEmailUser *model.User
	getByEmailErr  error

	createErr error

	existsByEmailCalls    int
	existsByUsernameCalls int
	getByEmailCalls       int
	createCalls           int

	createdUser *model.User
}

func (m *userRepoMock) ExistsByEmail(
	ctx context.Context,
	email string,
) (bool, error) {
	m.existsByEmailCalls++

	return m.existsByEmailResult, m.existsByEmailErr
}

func (m *userRepoMock) ExistsByUsername(
	ctx context.Context,
	username string,
) (bool, error) {
	m.existsByUsernameCalls++

	return m.existsByUsernameResult, m.existsByUsernameErr
}

func (m *userRepoMock) GetByEmail(
	ctx context.Context,
	email string,
) (*model.User, error) {
	m.getByEmailCalls++

	return m.getByEmailUser, m.getByEmailErr
}

func (m *userRepoMock) Create(
	ctx context.Context,
	user *model.User,
) error {
	m.createCalls++
	m.createdUser = user

	return m.createErr
}

func validUserCreateRequest() *model.UserCreateRequest {
	return &model.UserCreateRequest{
		Username: "testuser",
		Email:    "user@example.com",
		Password: "password123",
	}
}

func validUserLoginRequest() *model.UserLoginRequest {
	return &model.UserLoginRequest{
		Email:    "user@example.com",
		Password: "password123",
	}
}

func TestUserService_Register_InvalidRequest(t *testing.T) {
	repo := &userRepoMock{}
	service := NewUserService(repo, nil)

	req := &model.UserCreateRequest{}

	result, err := service.Register(context.Background(), req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.existsByEmailCalls != 0 {
		t.Fatal("repository must not be called for invalid request")
	}
}

func TestUserService_Register_InvalidEmail(t *testing.T) {
	repo := &userRepoMock{}
	service := NewUserService(repo, nil)

	req := validUserCreateRequest()
	req.Email = "invalid-email"

	result, err := service.Register(context.Background(), req)

	if err == nil {
		t.Fatal("expected email validation error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.existsByEmailCalls != 0 {
		t.Fatal("repository must not be called for invalid email")
	}
}

func TestUserService_Register_EmailAlreadyExists(t *testing.T) {
	repo := &userRepoMock{
		existsByEmailResult: true,
	}

	service := NewUserService(repo, nil)

	result, err := service.Register(
		context.Background(),
		validUserCreateRequest(),
	)

	if !errors.Is(err, blogerrors.ErrUserAlreadyExists) {
		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.existsByEmailCalls != 1 {
		t.Fatalf(
			"expected ExistsByEmail to be called once, got %d",
			repo.existsByEmailCalls,
		)
	}

	if repo.existsByUsernameCalls != 0 {
		t.Fatal("ExistsByUsername must not be called")
	}
}

func TestUserService_Register_EmailCheckError(t *testing.T) {
	repoErr := errors.New("database is unavailable")

	repo := &userRepoMock{
		existsByEmailErr: repoErr,
	}

	service := NewUserService(repo, nil)

	result, err := service.Register(
		context.Background(),
		validUserCreateRequest(),
	)

	// Согласно текущей реализации любая ошибка ExistsByEmail
	// преобразуется в ErrUserAlreadyExists.
	if !errors.Is(err, blogerrors.ErrUserAlreadyExists) {
		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUserService_Register_UsernameAlreadyExists(t *testing.T) {
	repo := &userRepoMock{
		existsByUsernameResult: true,
	}

	service := NewUserService(repo, nil)

	result, err := service.Register(
		context.Background(),
		validUserCreateRequest(),
	)

	if !errors.Is(err, blogerrors.ErrUserAlreadyExists) {
		t.Fatalf(
			"expected ErrUserAlreadyExists, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.existsByEmailCalls != 1 {
		t.Fatalf("expected email check to be called once")
	}

	if repo.existsByUsernameCalls != 1 {
		t.Fatalf("expected username check to be called once")
	}

	if repo.createCalls != 0 {
		t.Fatal("Create must not be called")
	}
}

func TestUserService_Register_UsernameCheckError(t *testing.T) {
	repoErr := errors.New("username query failed")

	repo := &userRepoMock{
		existsByUsernameErr: repoErr,
	}

	service := NewUserService(repo, nil)

	result, err := service.Register(
		context.Background(),
		validUserCreateRequest(),
	)

	if err == nil {
		t.Fatal("expected username check error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"ошибка при проверке имени пользователя",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !errors.Is(err, repoErr) {
		t.Fatalf(
			"expected wrapped repository error, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUserService_Register_UsernameNotFound(t *testing.T) {
	repoErr := errors.New("create failed")

	repo := &userRepoMock{
		existsByUsernameErr: sql.ErrNoRows,
		createErr:           repoErr,
	}

	service := NewUserService(repo, nil)

	result, err := service.Register(
		context.Background(),
		validUserCreateRequest(),
	)

	// sql.ErrNoRows при проверке username считается допустимым
	// результатом — пользователь с таким именем не найден.
	if err == nil {
		t.Fatal("expected create error, got nil")
	}

	if !strings.Contains(
		err.Error(),
		"ошибка при сохранении пользователя",
	) {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if !errors.Is(err, repoErr) {
		t.Fatalf(
			"expected wrapped create error, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.createCalls != 1 {
		t.Fatalf(
			"expected Create to be called once, got %d",
			repo.createCalls,
		)
	}

	if repo.createdUser == nil {
		t.Fatal("expected created user to be passed to repository")
	}

	if repo.createdUser.Email != "user@example.com" {
		t.Fatalf(
			"expected email %q, got %q",
			"user@example.com",
			repo.createdUser.Email,
		)
	}

	if repo.createdUser.Password == "password123" {
		t.Fatal("password must be hashed before saving")
	}

	if repo.createdUser.Password == "" {
		t.Fatal("hashed password must not be empty")
	}
}

func TestUserService_Login_InvalidRequest(t *testing.T) {
	repo := &userRepoMock{}
	service := NewUserService(repo, nil)

	req := &model.UserLoginRequest{}

	result, err := service.Login(context.Background(), req)

	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.getByEmailCalls != 0 {
		t.Fatal("repository must not be called for invalid request")
	}
}

func TestUserService_Login_InvalidEmail(t *testing.T) {
	repo := &userRepoMock{}
	service := NewUserService(repo, nil)

	req := validUserLoginRequest()
	req.Email = "invalid-email"

	result, err := service.Login(context.Background(), req)

	if err == nil {
		t.Fatal("expected email validation error, got nil")
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.getByEmailCalls != 0 {
		t.Fatal("repository must not be called for invalid email")
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	repo := &userRepoMock{
		getByEmailErr: sql.ErrNoRows,
	}

	service := NewUserService(repo, nil)

	result, err := service.Login(
		context.Background(),
		validUserLoginRequest(),
	)

	if !errors.Is(err, blogerrors.ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if result != nil {

		t.Fatalf("expected nil result, got %+v", result)
	}

	if repo.getByEmailCalls != 1 {
		t.Fatalf("expected GetByEmail to be called once")
	}
}

func TestUserService_Login_RepositoryError(t *testing.T) {
	repoErr := errors.New("database error")

	repo := &userRepoMock{
		getByEmailErr: repoErr,
	}

	service := NewUserService(repo, nil)

	result, err := service.Login(
		context.Background(),
		validUserLoginRequest(),
	)

	if !errors.Is(err, blogerrors.ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUserService_Login_NilUser(t *testing.T) {
	repo := &userRepoMock{
		getByEmailUser: nil,
	}

	service := NewUserService(repo, nil)

	result, err := service.Login(
		context.Background(),
		validUserLoginRequest(),
	)

	if !errors.Is(err, blogerrors.ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	hashedPassword, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &userRepoMock{
		getByEmailUser: &model.User{
			ID:       1,
			Username: "testuser",
			Email:    "user@example.com",
			Password: hashedPassword,
		},
	}

	service := NewUserService(repo, nil)

	req := &model.UserLoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	}

	result, err := service.Login(context.Background(), req)

	if !errors.Is(err, blogerrors.ErrInvalidCredentials) {
		t.Fatalf(
			"expected ErrInvalidCredentials, got %v",
			err,
		)
	}

	if result != nil {
		t.Fatalf("expected nil result, got %+v", result)
	}
}

func TestUserService_Login_GetByEmailCalledOnce(t *testing.T) {
	hashedPassword, err := auth.HashPassword("correct-password")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	repo := &userRepoMock{
		getByEmailUser: &model.User{
			ID:       1,
			Username: "testuser",
			Email:    "user@example.com",
			Password: hashedPassword,
		},
	}

	service := NewUserService(repo, nil)

	req := &model.UserLoginRequest{
		Email:    "user@example.com",
		Password: "wrong-password",
	}

	_, _ = service.Login(context.Background(), req)

	if repo.getByEmailCalls != 1 {
		t.Fatalf(
			"expected GetByEmail to be called once, got %d",
			repo.getByEmailCalls,
		)
	}
}
