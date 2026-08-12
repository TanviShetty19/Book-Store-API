package router

import (
	"log"
	"net/http"
	"time"

	"bookstore-api/internal/handler"
)

// statusRecorder wraps http.ResponseWriter to capture response status codes
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

// loggingMiddleware records incoming HTTP request logs and durations
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rec := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		log.Printf("[HTTP] %s %s | Status: %d %s | Duration: %v",
			r.Method,
			r.URL.Path,
			rec.statusCode,
			http.StatusText(rec.statusCode),
			duration,
		)
	})
}

// NewRouter registers all application routes with Carl (handler) and attaches middleware
func NewRouter(h *handler.BookHandler) http.Handler {
	mux := http.NewServeMux()

	// Register endpoints to Carl's handler methods
	mux.HandleFunc("GET /books", h.GetAllBooks)
	mux.HandleFunc("GET /books/{id}", h.GetBookByID)
	mux.HandleFunc("POST /books", h.CreateBook)
	mux.HandleFunc("PUT /books/{id}", h.UpdateBook)
	mux.HandleFunc("DELETE /books/{id}", h.DeleteBook)

	// Wrap entire mux with logging middleware
	return loggingMiddleware(mux)
}