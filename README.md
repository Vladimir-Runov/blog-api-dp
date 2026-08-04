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

## Технологический стек

### Инструменты

· Go 1.23+ — основной язык разработки.  
· PostgreSQL — реляционная база данных.  
· Docker + Docker Compose 

· Postman / curl — тестирование API.  
· GitHub / GitLab / GitVerse — репозиторий.  
· IDE VS Code + плагины Go.  

### Библиотеки  
· github.com/lib/pq — драйвер PostgreSQL  
· github.com/golang-jwt/jwt/v5 — JWT  
· golang.org/x/crypto/bcrypt — хеширование паролей  
· github.com/joho/godotenv — работа с .env  
· github.com/go-chi/chi/v5 или github.com/gin-gonic/gin — маршрутизация  
· github.com/jmoiron/sqlx — удобная работа с SQL  

### Технологический стек go
- **Язык:** Go 1.24
- **Аутентификация:** JWT (golang-jwt)


## Начало работы
### 1. Подготовка окружения
```bash
### Клонировать репозиторий и перейти в его папку.
git clone https://github.com/Vladimir-Runov/blog-api-dp.git
cd blog-api-dp
### Скопировать конфигурацию и прописать свои значения
cp .env.example .env
### Установить Go зависимости
go mod download
```
```bash
###2. Запустить docker PostgreSQL
docker-compose up -d

Подождать пока БД запустится.   
Проверить что БД работает на порту 5432: 'listening on IPv4 address "0.0.0.0", port 5432' (опционально)
docker-compose logs postgres
```

## Запуск  
по умолчанию, после запуска сервер будет доступен на `http://localhost:8080`  
```bash
### Запустить приложение 
go run cmd/api/main.go
### Запустить приложение с race detector 
go run -race ./cmd/api/main.go
### собрать и запустить
go build -o api ./cmd/api/main.go
./api
```




### 4. полезные команды Docker контейнеризация
```bash
### Запустить всё через Docker Compose
docker-compose up --build
### Остановить
docker-compose down
### Очистить данные БД
docker-compose down -v
```

## Разработка
### Публичные адреса
```
GET    /api/health                     # Проверка здоровья приложения  
POST   /api/register                 # Регистрация нового пользователя  
POST   /api/login                      # Вход пользователя  
GET    /api/posts                       # Получить все посты  
GET    /api/posts/{id}                 # Получить пост по ID
GET    /api/posts/{postId}/comments    # Получить комментарии к посту
```

### Защищенные адреса (требуют Authorization: Bearer <token>)
```
POST   /api/posts                      # Создать пост
PUT    /api/posts/{id}                 # Обновить пост
DELETE /api/posts/{id}                 # Удалить пост
POST   /api/posts/{postId}/comments    # Добавить комментарий
PUT    /api/comments/{id}              # Обновить комментарий
DELETE /api/comments/{id}              # Удалить комментарий
```




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

