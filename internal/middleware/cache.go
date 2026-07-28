package middleware

import (
	"bytes"
	"net/http"
	"sync"
	"time"
)

// ResponseCache хранит закешированный ответ
type ResponseCache struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// CacheItem хранит ответ и время истечения
type CacheItem struct {
	Value      *ResponseCache
	Expiration time.Time
}

// Cache — простой in-memory кеш
type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
	ttl   time.Duration
}

// NewCache создает новый кеш с TTL
func NewCache(ttl time.Duration) *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
		ttl:   ttl,
	}

	// Очистка устаревших элементов каждые 5 минут
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cache.mu.Lock()
			now := time.Now()
			for key, item := range cache.items {
				if now.After(item.Expiration) {
					delete(cache.items, key)
				}
			}
			cache.mu.Unlock()
		}
	}()

	return cache
}

// Get получает закешированный ответ
func (c *Cache) Get(key string) (*ResponseCache, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil, false
	}

	if time.Now().After(item.Expiration) {
		return nil, false
	}

	return item.Value, true
}

// Set сохраняет ответ в кеш
func (c *Cache) Set(key string, value *ResponseCache) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(c.ttl),
	}
}

// CacheMiddleware управляет кешированием GET-запросов
type CacheMiddleware struct {
	cache *Cache
}

// NewCacheMiddleware - создает новый middleware для кеширования
func NewCacheMiddleware(ttl time.Duration) *CacheMiddleware {
	return &CacheMiddleware{
		cache: NewCache(ttl),
	}
}

// CacheResponse кеширует только GET-запросы
func (m *CacheMiddleware) CacheResponse(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		// Формируем ключ кеша
		cacheKey := r.URL.Path + "?" + r.URL.RawQuery

		// Пытаемся получить из кеша
		if cached, found := m.cache.Get(cacheKey); found {
			for k, v := range cached.Headers {
				w.Header()[k] = v
			}
			w.WriteHeader(cached.StatusCode)
			w.Write(cached.Body)
			return
		}

		// Оборачиваем ResponseWriter
		ww := &cachingResponseWriter{
			ResponseWriter: w,
			body:           bytes.NewBuffer(nil),
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(ww, r)

		// Сохраняем в кеш только успешные ответы
		if ww.statusCode == http.StatusOK {
			m.cache.Set(cacheKey, &ResponseCache{
				StatusCode: ww.statusCode,
				Headers:    cloneHeader(ww.Header()),
				Body:       ww.body.Bytes(),
			})
		}
	})
}

// cachingResponseWriter буферизует ответ
type cachingResponseWriter struct {
	http.ResponseWriter
	statusCode int
	body       *bytes.Buffer
}

func (w *cachingResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *cachingResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func cloneHeader(h http.Header) http.Header {
	hh := http.Header{}
	for k, v := range h {
		hh[k] = make([]string, len(v))
		copy(hh[k], v)
	}
	return hh
}
