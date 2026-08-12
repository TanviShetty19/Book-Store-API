package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/service"
)

// statusRecorder wraps http.ResponseWriter to capture the HTTP status code for logging
type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (rec *statusRecorder) WriteHeader(code int) {
	rec.statusCode = code
	rec.ResponseWriter.WriteHeader(code)
}

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

func main() {
	//Initialize Repository (Sam)
	repo := repository.NewJSONBookRepository("data/books.json")
	//Initialize Service (Bob)
	svc := service.NewBookService(repo)
	//Initialize Handler (Carl)
	h := handler.NewBookHandler(svc)
	//Initialize HTTP Router (mux)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /books", h.GetAllBooks)
	mux.HandleFunc("GET /books/{id}", h.GetBookByID)
	mux.HandleFunc("POST /books",h.CreateBook)
	mux.HandleFunc("PUT /books/{id}", h.UpdateBook)
	mux.HandleFunc("DELETE /books/{id}",h.DeleteBook)

	wrappedMux := loggingMiddleware(mux)
	//Start Server
	fmt.Println("Bookstore API server running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", wrappedMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}