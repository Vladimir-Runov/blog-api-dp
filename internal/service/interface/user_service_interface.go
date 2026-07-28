package service

import (
	"blog-api-dp/internal/model"
	"context"
)

// UserServiceInterface определяет интерфейс для UserService, чтобы можно было мокировать его в тестах
type UserServiceInterface interface {
	Register(ctx context.Context, req *model.UserCreateRequest) (*model.TokenResponse, error)
	Login(ctx context.Context, req *model.UserLoginRequest) (*model.TokenResponse, error)
	GetByID(ctx context.Context, id int) (*model.User, error)
}
