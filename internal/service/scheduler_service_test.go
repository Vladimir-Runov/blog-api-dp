package service

import (
	"blog-api-dp/internal/model"
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockPostRepo - mock реализация PostRepository для тестирования
type mockPostRepo struct {
	mu                      sync.Mutex
	getScheduledPostsCalled bool
	getScheduledPostsReturn []*model.Post
	getScheduledPostsError  error
	callCount               int // Счетчик вызовов GetScheduledPosts
	publishPostCalled       []int
	publishPostError        error
}

func (m *mockPostRepo) Create(ctx context.Context, post *model.Post) error {
	return nil
}

func (m *mockPostRepo) GetByID(ctx context.Context, id int) (*model.Post, error) {
	return nil, nil
}

func (m *mockPostRepo) GetAll(ctx context.Context, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}

func (m *mockPostRepo) GetTotalCount(ctx context.Context) (int, error) {
	return 0, nil
}

func (m *mockPostRepo) Update(ctx context.Context, post *model.Post) error {
	return nil
}

func (m *mockPostRepo) Delete(ctx context.Context, id int) error {
	return nil
}

func (m *mockPostRepo) Exists(ctx context.Context, id int) (bool, error) {
	return false, nil
}

func (m *mockPostRepo) GetByAuthorID(ctx context.Context, authorID int, limit, offset int) ([]*model.Post, error) {
	return nil, nil
}
func (m *mockPostRepo) CountByAuthorID(ctx context.Context, authorID int) (int, error) {
	return 0, nil
}

func (m *mockPostRepo) GetTotalCountByAuthorID(ctx context.Context, authorID int) (int, error) {
	return 0, nil
}

func (m *mockPostRepo) GetScheduledPosts(ctx context.Context) ([]*model.Post, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getScheduledPostsCalled = true
	m.callCount++
	// Возвращаем посты только при первом вызове, симулируя публикацию
	if m.callCount == 1 {
		return m.getScheduledPostsReturn, m.getScheduledPostsError
	}
	return []*model.Post{}, m.getScheduledPostsError
}

func (m *mockPostRepo) PublishPost(ctx context.Context, id int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishPostCalled = append(m.publishPostCalled, id)
	return m.publishPostError
}

func TestSchedulerService_Start_Stop_NoPosts(t *testing.T) {
	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: []*model.Post{},
	}

	scheduler := NewSchedulerService(mockRepo, 1, 100*time.Millisecond)
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(200 * time.Millisecond) // Подождем немного, чтобы ticker сработал
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
	assert.Empty(t, mockRepo.publishPostCalled)
}

func TestSchedulerService_Start_Stop_WithPosts(t *testing.T) {
	now := time.Now()
	posts := []*model.Post{
		{ID: 1, Title: "Post 1", Status: "draft", PublishAt: &now},
		{ID: 2, Title: "Post 2", Status: "draft", PublishAt: &now},
	}

	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: posts,
	}

	scheduler := NewSchedulerService(mockRepo, 2, 100*time.Millisecond) // 2 воркера
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(200 * time.Millisecond) // Подождать публикацию
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
	assert.Len(t, mockRepo.publishPostCalled, 2)
	assert.Contains(t, mockRepo.publishPostCalled, 1)
	assert.Contains(t, mockRepo.publishPostCalled, 2)
}

func TestSchedulerService_Start_Stop_PublishError(t *testing.T) {
	now := time.Now()
	posts := []*model.Post{
		{ID: 1, Title: "Post 1", Status: "draft", PublishAt: &now},
	}

	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: posts,
		publishPostError:        errors.New("publish error"),
	}

	scheduler := NewSchedulerService(mockRepo, 1, 100*time.Millisecond)
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
	assert.Len(t, mockRepo.publishPostCalled, 1)
	assert.Contains(t, mockRepo.publishPostCalled, 1)
}

func TestSchedulerService_Start_Stop_GetScheduledPostsError(t *testing.T) {
	mockRepo := &mockPostRepo{
		getScheduledPostsError: errors.New("get scheduled posts error"),
	}

	scheduler := NewSchedulerService(mockRepo, 1, 100*time.Millisecond)
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
	assert.Empty(t, mockRepo.publishPostCalled)
}

func TestSchedulerService_Start_Stop_ContextCancelled(t *testing.T) {
	now := time.Now()
	posts := []*model.Post{
		{ID: 1, Title: "Post 1", Status: "draft", PublishAt: &now},
	}

	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: posts,
	}

	scheduler := NewSchedulerService(mockRepo, 1, 10*time.Millisecond) // Уменьшенный poll interval - для гарантированного вызова
	ctx, cancel := context.WithCancel(context.Background())

	scheduler.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)
	scheduler.Stop()

	// Публикация может произойти или нет (в зависимости от тайминга)
	assert.True(t, mockRepo.getScheduledPostsCalled)
}

func TestSchedulerService_Stop_WithoutStart(t *testing.T) {
	mockRepo := &mockPostRepo{}
	scheduler := NewSchedulerService(mockRepo, 1, 100*time.Millisecond)

	// Stop без Start не должен паниковать
	scheduler.Stop()
}

func TestSchedulerService_Start_MultiplePosts_ConcurrentWorkers(t *testing.T) {
	now := time.Now()
	posts := []*model.Post{
		{ID: 1, Title: "Post 1", Status: "draft", PublishAt: &now},
		{ID: 2, Title: "Post 2", Status: "draft", PublishAt: &now},
		{ID: 3, Title: "Post 3", Status: "draft", PublishAt: &now},
	}

	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: posts,
	}

	scheduler := NewSchedulerService(mockRepo, 2, 100*time.Millisecond) // 2 воркера
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(200 * time.Millisecond)
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
	assert.Len(t, mockRepo.publishPostCalled, 3)
	assert.Contains(t, mockRepo.publishPostCalled, 1)
	assert.Contains(t, mockRepo.publishPostCalled, 2)
	assert.Contains(t, mockRepo.publishPostCalled, 3)
}

func TestSchedulerService_Start_LongRunning_NoImmediateStop(t *testing.T) {
	mockRepo := &mockPostRepo{
		getScheduledPostsReturn: []*model.Post{},
	}

	scheduler := NewSchedulerService(mockRepo, 1, 50*time.Millisecond) // Уменьшенный интервал для теста
	ctx := context.Background()

	scheduler.Start(ctx)
	time.Sleep(100 * time.Millisecond) // Время, чтобы ticker сработал
	scheduler.Stop()

	assert.True(t, mockRepo.getScheduledPostsCalled)
}
