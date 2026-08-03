# blog-api-dp
Дипломная работа профессии «Go-разработчик 2026» 
Рунов В.В.

## 📋 Содержание

- [О проекте](#о-проекте)
- [Структура проекта](#структура-проекта)
- [Технологический стек](#технологический-стек)
- [Быстрый старт](#быстрый-старт)
- [Разработка](#разработка)
- [API эндпоинты](#api-эндпоинты)
- [Планировщик отложенной публикации](#планировщик-отложенной-публикации)
- [Тестирование](#тестирование)
- [Документация](#документация)
- [Архитектура приложения](#архитектура-приложения)

## О проекте

Расширенная система управления блогом на Go.

Проект включает:
- ✅ Управление пользователями (регистрация, вход)
- ✅ JWT аутентификация и авторизация
- ✅ Управление постами (создание, редактирование, удаление, контроль принадлежности)
- ✅ Управление комментариями (создание, редактирование, удаление, контроль принадлежности)
в т.ч.
- ✅ Отложенная публикация постов 
- ✅ Логирование всех запросов с уникальными request_id
- ✅ Docker контейнеризация
- ✅ Покрытие тестами

## Структура проекта
```
-blog-api-dp/
├── cmd/api/
│   └── main.go                 # Точка входа приложения
├── internal/
├── config/              # 
│   │   └── config.go
│   ├── errors/              # Ошибки приложения
│   │   └── errors.go
│   ├── handler/                # HTTP обработчики
│   │   ├── auth_handler.go
│   │   ├── comment_handler.go
│   │   ├── post_handler.go
│   │   └── *_test.go
│   ├── middleware/             # HTTP middleware
│   │   ├── auth.go
│   │   ├── cache.go
│   │   ├── logging.go
│   │   ├── recovery.go
│   │   └── *_test.go
│   ├── model/                  # Модели данных
│   │   ├── models.go
│   │   └── *_test.go
│   ├── repository/             # Слой доступа к БД
│   │   ├── comments_repo.go
│   │   ├── comments_repo_interface.go
│   │   ├── post_repo.go
│   │   ├── post_repo_interface.go
│   │   ├── user_repo.go
│   │   ├── user_repo_interface.go
│   │   └── *repo_test.go
│   └── service/                # Бизнес-логика
│       ├── comment_service.go
│       ├── post_service.go
│       ├── scheduler.go
│       ├── user_service.go
│       └── *_service_test.go
│       ├── interface/             # интерфейс сервисов
├── migrations/                 # Postgres SQL миграции
│   ├── 001_init_schema.sql
│   ├── 002_indexes.sql
├── pkg/
│   ├── auth/                   # Утилиты аутентификации
│   │   ├── jwt.go
│   │   ├── password.go
│   │   └── *_test.go
│   └── database/               # Утилиты БД
│       ├── init.go
│       ├── migrate.go
├── .env.example                # Пример конфигурации
├── docker-compose.yml          # Docker Compose
├── Dockerfile                  # Docker образ
├── go.mod                      # Зависимости проекта
└── README.md
```




## Запустить docker PostgreSQL
docker-compose up -d
## Проверить что БД работает на порту 5432: 'listening on IPv4 address "0.0.0.0", port 5432' (опционально)
docker-compose logs postgres



## запуск  
go run cmd/api/main.go

## Остановить  
docker-compose down  
## Очистить данные БД  
docker-compose down -v  

## Тестирование

### Unit тесты

```bash
# Запустить все тесты
go test ./...

# Запустить с подробным выводом
go test -v ./...

# Проверить c race conditions
go test -race ./...

# Посмотреть покрытие тестами
go test ./... -cover
```

