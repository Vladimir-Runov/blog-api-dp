package middleware

import (
	"context"
	"log"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LoggingMiddleware provides request logging, CORS, recovery and other utility middleware
type LoggingMiddleware struct {
	logger *log.Logger
}

// responseWriter обертка для захвата статус кода
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// WriteHeader переопределяет метод WriteHeader для захвата статус кода
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.ResponseWriter.WriteHeader(code)
		rw.written = true
	}
}

// NewLoggingMiddleware creates a new logging middleware instance
func NewLoggingMiddleware(logger *log.Logger) *LoggingMiddleware {
	return &LoggingMiddleware{
		logger: logger,
	}
}

// Logger logs all HTTP requests
func (m *LoggingMiddleware) Logger(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Засечь время начала запроса
		start := time.Now()

		// 2. Создать wrapper для ResponseWriter чтобы захватить статус код
		wrappedWriter := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		// 3. Вызвать следующий handler с wrapped writer
		next(wrappedWriter, r)

		// 4. После выполнения залогировать: метод, путь, IP, статус, время выполнения
		m.logger.Printf("%s %s %s %d %s", r.Method, r.URL.Path, r.RemoteAddr, wrappedWriter.statusCode, time.Since(start))
	}
}

// Recovery восстанавливается после паник
func (m *LoggingMiddleware) Recovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// 1. Логируем ошибку
				m.logger.Printf("Recovered from panic: %v", err)

				// 2. Опционально: добавляем stack trace
				buf := make([]byte, 1024)
				n := runtime.Stack(buf, true)
				m.logger.Printf("Stack trace:\n%s", buf[:n])

				// 3. Возвращаем клиенту 500 Internal Server Error
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		// 4. Вызываем следующий handler
		next(w, r)
	}
}

// CORS добавляет CORS заголовки
func (m *LoggingMiddleware) CORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Добавляем необходимые CORS заголовки
		w.Header().Set("Access-Control-Allow-Origin", "*") // Разрешаем все источники (можно указать конкретный)
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400") // Кэшируем preflight запросы на 1 день

		// 2. Обрабатываем preflight запросы (OPTIONS метод)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // Возвращаем 204 No Content
			return
		}

		// 3. Для остальных методов вызываем следующий handler
		next(w, r)
	}
}

// RequestID добавляет уникальный ID к каждому запросу
func (m *LoggingMiddleware) RequestID(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Генерируем уникальный ID (UUID)
		requestID := uuid.New().String()

		// 2. Добавляем ID в контекст запроса
		ctx := context.WithValue(r.Context(), "RequestID", requestID)
		r = r.WithContext(ctx)

		// 3. Добавляем ID в заголовок ответа X-Request-ID
		w.Header().Set("X-Request-ID", requestID)

		// 4. Логируем запрос с Request ID
		m.logger.Printf("Received request %s %s with Request ID: %s", r.Method, r.URL.Path, requestID)

		// 5. Вызываем следующий обработчик
		next(w, r)
	}
}

// RateLimiter ограничивает количество запросов от одного клиента
func (m *LoggingMiddleware) RateLimiter(maxRequests int, window time.Duration) func(http.HandlerFunc) http.HandlerFunc {
	// TODO: Реализовать rate limiting (продвинутое задание)
	// Шаги:
	// 1. Создать хранилище для отслеживания запросов по IP адресам
	// 2. Использовать mutex для безопасного доступа к хранилищу
	// 3. Для каждого запроса:
	//    - Получить IP клиента
	//    - Проверить количество запросов в текущем окне времени
	//    - Если превышен лимит - вернуть 429 Too Many Requests
	//    - Иначе увеличить счетчик и пропустить запрос
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {

			// Временная реализация
			next(w, r)
		}
	}
}

// ContentTypeJSON устанавливает Content-Type: application/json для всех ответов
func (m *LoggingMiddleware) ContentTypeJSON(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Установить Content-Type: application/json для всех ответов
		w.Header().Set("Content-Type", "application/json")

		// Вызываем следующий обработчик
		next(w, r)
	}
}

// getClientIP извлекает IP адрес клиента
// позволяет корректно извлекать IP-адрес клиента в различных сценариях, включая использование прокси-серверов.
func getClientIP(r *http.Request) string {
	// TODO: Извлечь реальный IP адрес клиента
	// Проверить заголовки: X-Forwarded-For, X-Real-IP, затем RemoteAddr
	// Учесть что X-Forwarded-For может содержать несколько IP
	//return r.RemoteAddr
	// Проверяем заголовок X-Forwarded-For
	//– Сначала проверяется заголовок X-Forwarded-For. Если он не пустой, то предполагается, что он может содержать несколько IP-адресов, разделенных запятыми. Мы берем первый IP и убираем лишние пробелы.
	//– Затем проверяется заголовок X-Real-IP. Если он установлен, то возвращается его значение.
	//
	//	2. Использование RemoteAddr:
	//	– Если оба заголовка отсутствуют, используется RemoteAddr, который содержит IP-адрес клиента. Функция net.SplitHostPort разбивает строку на IP и порт. Если возникает ошибка при разбиении, возвращается оригинальное значение RemoteAddr.
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	if xForwardedFor != "" {
		// X-Forwarded-For может содержать несколько IP, берем первый
		ips := strings.Split(xForwardedFor, ",")
		return strings.TrimSpace(ips[0])
	}

	// Проверяем заголовок X-Real-IP
	xRealIP := r.Header.Get("X-Real-IP")
	if xRealIP != "" {
		return xRealIP
	}

	// используем RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr // Возвращаем RemoteAddr в случае ошибки
	}
	return ip
}

// Write вызывает WriteHeader если еще не был вызван
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(b)
}

// newResponseWriter создает новую обертку
func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		written:        false,
	}
}
