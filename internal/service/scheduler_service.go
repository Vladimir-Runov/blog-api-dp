package service

import (
	"blog-api-dp/internal/model"
	"blog-api-dp/internal/repository"
	"context"
	"log"
	"sync"
	"time"
)

// SchedulerService сервис для планирования отложенных публикаций
type SchedulerService struct {
	postRepo     repository.PostRepository
	workerCount  int
	pollInterval time.Duration
	wg           sync.WaitGroup
	stopChan     chan struct{}
}

// NewSchedulerService создает новый сервис планировщика
func NewSchedulerService(postRepo repository.PostRepository, workerCount int, pollInterval time.Duration) *SchedulerService {
	return &SchedulerService{
		postRepo:     postRepo,
		workerCount:  workerCount,
		pollInterval: pollInterval,
		stopChan:     make(chan struct{}),
	}
}

// Start запускает планировщик
func (s *SchedulerService) Start(ctx context.Context) {
	log.Printf("Starting scheduler with %d workers, polling every %v", s.workerCount, s.pollInterval)

	// Создаем worker pool
	jobs := make(chan *model.Post, s.workerCount*2)

	// Запускаем workers
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)
		go s.worker(ctx, i+1, jobs)
	}

	// Запускаем планировщик
	s.wg.Add(1)
	go s.scheduler(ctx, jobs)
}

// Stop останавливает планировщик
func (s *SchedulerService) Stop() {
	log.Println("Stopping scheduler...")
	close(s.stopChan)
	s.wg.Wait()
	log.Println("Scheduler stopped")
}

// publishPost публикует пост
func (s *SchedulerService) publishPost(ctx context.Context, post *model.Post, workerID int) {
	log.Printf("Worker %d processing post %d: %s", workerID, post.ID, post.Title)

	if err := s.postRepo.PublishPost(ctx, post.ID); err != nil {
		log.Printf("Worker %d failed to publish post %d: %v", workerID, post.ID, err)
		return
	}

	log.Printf("Worker %d successfully published post %d: %s", workerID, post.ID, post.Title)
}

// scheduler основной цикл планировщика
func (s *SchedulerService) scheduler(ctx context.Context, jobs chan<- *model.Post) {
	defer s.wg.Done()
	defer close(jobs)

	ticker := time.NewTicker(s.pollInterval)
	defer ticker.Stop()

	log.Println("\tScheduler ticker")
	for {
		select {
		case <-s.stopChan:
			log.Println("\tScheduler received stop signal")
			return
		case <-ticker.C:
			log.Println("\tScheduler ticker processScheduledPosts ...")
			s.processScheduledPosts(ctx, jobs)
		case <-ctx.Done():
			log.Println("Scheduler context cancelled")
			return
		}
	}
}

// worker обрабатывает посты для публикации
func (s *SchedulerService) worker(ctx context.Context, id int, jobs <-chan *model.Post) {
	defer s.wg.Done()

	log.Printf("Worker %d started", id)

	for post := range jobs {
		log.Println("\t worker post ...")
		select {
		case <-s.stopChan:
			log.Printf("Worker %d received stop signal", id)
			return
		case <-ctx.Done():
			log.Printf("Worker %d context cancelled", id)
			return
		default:
			s.publishPost(ctx, post, id)
		}
	}

	log.Printf("Worker %d stopped", id)
}

// processScheduledPosts получает посты, которые готовы к публикации
func (s *SchedulerService) processScheduledPosts(ctx context.Context, jobs chan<- *model.Post) {
	posts, err := s.postRepo.GetScheduledPosts(ctx)

	if err != nil {
		log.Printf("Error getting scheduled posts: %v", err)
		return
	}

	if len(posts) > 0 {
		log.Printf("  !Found %d posts ready for publication", len(posts))
	} else {
		log.Printf(" a posts to schedule were not found")
	}

	for _, post := range posts {
		select {
		case jobs <- post:
			log.Printf("put post %d for publication", post.ID)
		case <-s.stopChan:
			log.Println("Stopping scheduling due to shutdown signal")
			return
		case <-ctx.Done():
			log.Println("Stopping scheduling due to context cancellation")
			return
		}
	}
}
