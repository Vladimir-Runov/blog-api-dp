package main

// c:\Users\Admin\Documents\go\git_netology\blog-api-dp
//  go test ./...

import (
	"blog-api-dp/internal/config"
	"blog-api-dp/internal/handler"
	middlewareauth "blog-api-dp/internal/middleware"
	"blog-api-dp/internal/repository"
	"blog-api-dp/internal/service"
	"blog-api-dp/pkg/auth"
	"blog-api-dp/pkg/database"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"context"
	"syscall"

	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Загружаем конфигурацию из .env файла
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found, using default environment variables")
	}

	cfg := config.EnvloadConfig() //  конфигурацию из переменных окружения
	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.DBHost,
		Port:     cfg.DBPort,
		User:     cfg.DBUser,
		Password: cfg.DBPassword,
		DBName:   cfg.DBName,
		SSLMode:  cfg.DBSSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close(db)

	log.Println("Running database migrations...")
	if err := database.Migrate(db, false); err != nil { //  true - to delete schema (reset database structure)
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed successfully")

	userRepo := repository.NewUserRepo(db)
	postRepo := repository.NewPostRepo(db)
	commentRepo := repository.NewCommentRepo(db)

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiryHours)

	userService := service.NewUserService(userRepo, jwtManager)
	postService := service.NewPostService(postRepo, userRepo, jwtManager)
	commentService := service.NewCommentService(commentRepo, postRepo, userRepo)

	userHandler := handler.NewAuthHandlerEx(userService)
	postHandler := handler.NewPostHandler(postService)
	commentHandler := handler.NewCommentHandler(commentService)

	log.Printf("Initializing scheduler with %d workers, polling every %v", 5, 30*time.Second)
	schedulerService := service.NewSchedulerService(postRepo,
		5,              // worker count
		30*time.Second, // poll interval
	)

	router := chi.NewRouter()

	router.Use(middleware.Throttle(100)) // Ограничение на 100 одновременных запросов
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.AllowContentType("application/json")) // CORS middleware

	// Health check эндпоинт
	router.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok","service":"blog-api"}`))
	})

	router.Post("/api/register", userHandler.Register) // - POST /api/register
	router.Post("/api/login", userHandler.Login)       // - POST /api/login

	router.Get("/api/posts", postHandler.GetAll)                         // - GET /api/posts
	router.Get("/api/posts/{id}", postHandler.GetByID)                   // - GET /api/posts/{id}
	router.Get("/api/comments/{id}", commentHandler.GetByID)             // - GET /api/comments/{id}
	router.Get("/api/posts/{postId}/comments", commentHandler.GetByPost) //	· получение комментариев к посту (GET /api/posts/{postId}/comments) — доступно всем.

	authMiddleware := middlewareauth.NewAuthMiddleware(jwtManager) // middlewareauth “blog-example-go-restapi/internal/middleware”

	router.Post("/api/posts", authMiddleware.RequireAuth(postHandler.Create))        // - POST /api/posts (требует JWT) создание поста (POST /api/posts) — только для авторизованных;
	router.Put("/api/posts/{id}", authMiddleware.RequireAuth(postHandler.Update))    // - PUT /api/posts/{id} (требует JWT) обновление поста (PUT /api/posts/{id}) — только автор;
	router.Delete("/api/posts/{id}", authMiddleware.RequireAuth(postHandler.Delete)) // - DELETE /api/posts/{id} (требует JWT)

	router.Post("/api/posts/{postId}/comments", authMiddleware.RequireAuth(commentHandler.Create)) // создание комментария по заданию должно идти через путь /api/posts/{postId}/comments
	router.Put("/api/posts/{postId}/comments", authMiddleware.RequireAuth(commentHandler.Update))  //

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Запуск планировщика
	log.Println("Starting scheduler...")
	schedulerService.Start(ctx)

	addr := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Starting server on %s...", addr)
	// - Запустить сервер и обработать ошибки
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}

	server := &http.Server{
		Addr:    cfg.ServerHost + ":" + strconv.Itoa(cfg.ServerPort),
		Handler: router,
	}

	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Ожидание сигналов для graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Остановка планировщика
	log.Println("Stopping scheduler...")
	schedulerService.Stop()

	// Graceful shutdown сервера
	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
