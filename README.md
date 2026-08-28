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
- ✅ Логирование всех запросов с уникальными id (в виде  20260827085244.044-500 ), что позволяет связать все события, относящиеся к одному HTTP-запросу.
в т.ч.
- ✅ Отложенная публикация постов 
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
│   │   ├── logging.go
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
Проверить что БД работает 
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

## Планировщик отложенной публикации  
1. При создании поста с указанием `publish_at` в будущем, пост сохраняется в статусе "draft"  
2. Планировщик каждые 30 секунд проверяет посты со статусом "draft" и временем публикации ранее текущего времени, такие  посты автоматически 'публикуются' (статус   меняется на "published", поле publish_at устанавливается в NULL)
4. Обработка публикации отложенных постов происходит конкурентно с использованием worker pool (5 воркеров по умолчанию)  
5. Все операции логируются для отслеживания процесса  
  
### Параметры планировщика (настраивается в коде)  
- Интервал проверки: 30 секунд   
- Количество воркеров: 5   
- Планировщик корректно завершается при остановке сервиса (graceful shutdown)  

### Пример использования. Создание поста с отложенной публикацией:
```bash
curl -X POST http://localhost:8080/api/posts \
  -H "Authorization: Bearer <токен>" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "заголовок отложенного поста",
    "content": "Этот пост будет опубликован в 13:22  01.08,2026,
    "publish_at": "2026-08-01T13:22:23Z"
  }'
```


## Тестирование

### Unit тесты

```bash
# Запустить все тесты
go test ./...

### Для подробного вывода:
go test -count=1 -v ./...   


### при последующих запусках можно предварительно очистить кэш или отключение его использования:
go clean -cache -testcache
go test -count=1 ./... 

### Запустить с подробным выводом
go test -v ./...


## Проверить c race conditions
go test -race ./...

# Посмотреть покрытие тестами
go test ./... -cover
```

/cmd/api             coverage: 0.0% of statements
/internal/config			3.294s  coverage: 100.0% of statements
/internal/erros			4.181s  coverage: 100.0% of statements
/internal/handler			2.923s  coverage: 19.9% of statements
/internal/middleware 			1.565s  coverage: 71.1% of statements
/internal/repository		1.272s  coverage: 62.2% of statements
/internal/service			4.434s  coverage: 63.0% of statements
/pkg/auth			1.335s  coverage: 41.2% of statements
/pkg/database			0.999s  coverage: 10.1% of statements

## Архитектура приложения

Проект использует чистую архитектуру с разделением ответственности:

```
┌─────────────────┐
│  HTTP Requests  │
└────────┬────────┘
         │
┌────────▼────────────────────────┐
│ Middleware (Auth, Logging, CORS)│
└────────┬────────────────────────┘
         │
┌────────▼─────────────┐
│ Handlers (HTTP API)  │ ← Парсинг JSON, валидация, HTTP коды
└────────┬─────────────┘
         │
┌────────▼──────────────┐
│ Services (Business)   │ ← Бизнес-логика, валидация, права
└────────┬──────────────┘
         │
┌────────▼─────────────┐
│ Repositories (Data)  │ ← SQL запросы, работа с БД
└────────┬─────────────┘
         │
┌────────▼────────┐
│  PostgreSQL DB  │
└─────────────────┘
```

### Docker образы

- **API:** golang:1.24-alpine → компиляция → alpine:latest (многоэтапная сборка)
- **База данных:** postgres:15
- **Сеть:** bridge с именем blog-network
- **Том:** postgres_data для сохранения данных

### Миграции

1. **001_init_schema.sql** - создаёт таблицы users, posts, comments
2. **002_index.sql** - добавляет индексы к созданным таблицам.

Миграции запускаются автоматически при старте приложения.

## Конфигурация
Переменные окружения задаются в файле `.env`:

```env
# Server Configuration
SERVER_PORT=8080
SERVER_HOST=localhost

# Database Configuration, port redirected !
DB_HOST=localhost
DB_PORT=5433
DB_USER=bloguser
DB_PASSWORD=postgres
DB_NAME=blogdb
DB_SSLMODE=disable

# JWT Configuration
JWT_SECRET=our-secret-key-change-in-production
JWT_EXPIRY_HOURS=2

# Application Configuration
APP_ENV=development
LOG_LEVEL=debug

# Cache Configuration
CACHE_TTL_MINUTES=5
```


## Полезные команды

```bash
# Скачать и обновить зависимости
go mod download
go mod tidy

# Запустить приложение
go run -race ./cmd/api/main.go

# Собрать приложение
go build -o api ./cmd/api/main.go

# Запустить тесты
go test ./... -v

# Просмотреть логи БД
docker-compose logs db -f

# Подключиться к БД
docker-compose exec db psql -U postgres -d blog_db

# Остановить все сервисы
docker-compose down

# Очистить данные БД
docker-compose down -v
```

## SQL запросы для ручного тестирования

```sql
-- Подключиться к БД
docker-compose exec db psql -U postgres -d blog_db

-- Просмотреть таблицы
\dt

-- Просмотреть пользователей
SELECT id, username, email, created_at FROM users;

-- Просмотреть посты
SELECT id, title, status, author_id, created_at FROM posts;

-- Просмотреть комментарии
SELECT id, content, post_id, author_id, created_at FROM comments;

-- Проверить посты в статусе draft
SELECT id, title, status, publish_at FROM posts WHERE status = 'draft';
```


