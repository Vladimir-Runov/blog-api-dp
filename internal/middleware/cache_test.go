package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewCache тестирует создание нового кеша с заданным TTL
func TestNewCache(t *testing.T) {
	// Создаем кеш с TTL 1 минута
	ttl := 1 * time.Minute
	cache := NewCache(ttl)

	// Проверяем, что кеш инициализирован
	assert.NotNil(t, cache)
	assert.Equal(t, ttl, cache.ttl)
	assert.NotNil(t, cache.items)
}

// TestCache_Get тестирует получение элементов из кеша
func TestCache_Get(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	// Устанавливаем элемент
	key := "test_key"
	value := &ResponseCache{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"status":"ok"}`),
	}
	cache.Set(key, value)

	// Получаем существующий элемент
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)

	// Получаем несуществующий элемент
	_, found = cache.Get("nonexistent")
	assert.False(t, found)

	// Тестируем истекший элемент
	expiredCache := NewCache(1 * time.Millisecond) // Очень короткий TTL
	expiredCache.Set(key, value)
	time.Sleep(2 * time.Millisecond) // Ждем истечения

	_, found = expiredCache.Get(key)
	assert.False(t, found)
}

// TestCache_Set тестирует установку элементов в кеш
func TestCache_Set(t *testing.T) {
	cache := NewCache(5 * time.Minute)

	key := "test_key"
	value := &ResponseCache{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"status":"ok"}`),
	}

	// Устанавливаем элемент
	cache.Set(key, value)

	// Проверяем, что элемент установлен
	retrieved, found := cache.Get(key)
	assert.True(t, found)
	assert.Equal(t, value, retrieved)
}

// TestCache_Expiration тестирует автоматическое удаление истекших элементов
// Этот тест проверяет механизм очистки (в реальности очистка происходит каждые 5 минут).
// Мы симулируем это вручную - для тестирования
func TestCache_Expiration(t *testing.T) {
	cache := NewCache(1 * time.Millisecond)

	key := "test_key"
	value := &ResponseCache{
		StatusCode: 200,
		Headers:    http.Header{"Content-Type": []string{"application/json"}},
		Body:       []byte(`{"status":"ok"}`),
	}

	// Устанавливаем элемент
	cache.Set(key, value)

	// Проверяем, что элемент есть
	_, found := cache.Get(key)
	assert.True(t, found)

	// Ждем истечения TTL
	time.Sleep(2 * time.Millisecond)

	// Симулируем очистку (что у нас делает тикер)
	cache.mu.Lock()
	now := time.Now()
	for k, item := range cache.items {
		if now.After(item.Expiration) {
			delete(cache.items, k)
		}
	}
	cache.mu.Unlock()

	// Проверяем, что элемент удален
	_, found = cache.Get(key)
	assert.False(t, found)
}

// TestCacheMiddleware_CacheResponse тестирует middleware кеширования
func TestCacheMiddleware_CacheResponse(t *testing.T) {
	// Создаем middleware с TTL 5 минут
	ttl := 5 * time.Minute
	cacheMiddleware := NewCacheMiddleware(ttl)

	// Создаем тестовый handler, который возвращает JSON
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Оборачиваем handler middleware
	cachedHandler := cacheMiddleware.CacheResponse(testHandler)

	// Тест 1: GET-запрос должен кешироваться
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	cachedHandler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, `{"status":"ok"}`, rec.Body.String())

	// Тест 2: POST-запрос не должен кешироваться
	reqPost := httptest.NewRequest("POST", "/api/test", bytes.NewReader([]byte{}))
	recPost := httptest.NewRecorder()

	cachedHandler.ServeHTTP(recPost, reqPost)

	assert.Equal(t, http.StatusOK, recPost.Code) // Handler все равно выполнится

	// Тест 3: Кеширование только успешных ответов (200 OK)
	// Создаем handler, который возвращает ошибку
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal server error"}`))
	})

	cachedErrorHandler := cacheMiddleware.CacheResponse(errorHandler)

	reqError := httptest.NewRequest("GET", "/api/error", nil)
	recError := httptest.NewRecorder()

	cachedErrorHandler.ServeHTTP(recError, reqError)

	assert.Equal(t, http.StatusInternalServerError, recError.Code)
}

// TestCloneHeader тестирует функцию cloneHeader
func TestCloneHeader(t *testing.T) {
	original := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer token"},
		"X-Custom":      []string{"value1", "value2"},
	}

	cloned := cloneHeader(original)

	// Проверяем, что заголовки идентичны
	assert.Equal(t, original, cloned)

	// Проверяем, что это разные объекты (глубокое копирование)
	original["Content-Type"][0] = "text/plain"
	assert.NotEqual(t, original["Content-Type"][0], cloned["Content-Type"][0])
}

// TestCachingResponseWriter_WriteHeader тестирует cachingResponseWriter
func TestCachingResponseWriter_WriteHeader(t *testing.T) {
	body := &bytes.Buffer{}
	writer := &cachingResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		body:           body,
		statusCode:     http.StatusOK,
	}

	// Устанавливаем статус код
	writer.WriteHeader(http.StatusNotFound)
	assert.Equal(t, http.StatusNotFound, writer.statusCode)
}

// TestCachingResponseWriter_Write тестирует запись в cachingResponseWriter
func TestCachingResponseWriter_Write(t *testing.T) {
	body := &bytes.Buffer{}
	writer := &cachingResponseWriter{
		ResponseWriter: httptest.NewRecorder(),
		body:           body,
		statusCode:     http.StatusOK,
	}

	data := []byte("test data")
	n, err := writer.Write(data)

	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, data, body.Bytes())
}
