package middleware

import (
	"net/http"
	"sync"
	"time"
)

// Throttle создает middleware с ограничением количества одновременных запросов
func Throttle(maxConcurrent int) func(http.Handler) http.Handler {
	sem := make(chan struct{}, maxConcurrent)
	mu := sync.RWMutex{}
	count := 0

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			count = 0
			mu.Unlock()
		}
	}()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.RLock()
			current := count
			mu.RUnlock()

			if current >= maxConcurrent*60 { // 100 req/min
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			sem <- struct{}{}        // acquire
			defer func() { <-sem }() // release

			mu.Lock()
			count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}
