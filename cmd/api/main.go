package main

import (
	"fmt"
	"log"
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/router"
	"bookstore-api/internal/service"
)

func main() {
	// 1. Initialize Repositories (Data Access Layer)
	bookRepo := repository.NewJSONBookRepository("books.json")
	userRepo := repository.NewMemoryUserRepository()

	// 2. Initialize Services (Business Logic Layer)
	bookService := service.NewBookService(bookRepo)
	authService := service.NewAuthService(userRepo)

	// 3. Initialize Handlers (Presentation/HTTP Layer)
	bookHandler := handler.NewBookHandler(bookService)
	authHandler := handler.NewAuthHandler(authService)

	// 4. Initialize Base Router
	appRouter := router.NewRouter(bookHandler, authHandler)

	// 5. Build Global Middleware Pipeline
	// Execution Flow: Recovery (Outermost) -> Logging -> CORS -> Router (Innermost)
	pipeline := middleware.Chain(
		appRouter,
		middleware.RecoveryMiddleware,
		middleware.LoggingMiddleware,
		middleware.CORSMiddleware,
	)

	fmt.Println("Bookstore API server running on http://localhost:8080...")
	log.Fatal(http.ListenAndServe(":8080", pipeline))
}