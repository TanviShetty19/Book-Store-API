package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/router"
	"bookstore-api/internal/service"
)

func main() {
	// 1. Enforce strict JWT secret requirement (No insecure default fallbacks)
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required; server refusing to start without it")
	}

	// 2. Initialize Repositories (Passing shared bookRepo to OrderRepo)
	bookRepo, err := repository.NewJsonBookRepository("data/books.json")
	if err != nil {
		log.Fatalf("Failed to initialize book repository: %v", err)
	}

	userRepo, err := repository.NewJsonUserRepository("data/users.json")
	if err != nil {
		log.Fatalf("Failed to initialize user repository: %v", err)
	}

	orderRepo, err := repository.NewJsonOrderRepository("data/orders.json", bookRepo)
	if err != nil {
		log.Fatalf("Failed to initialize order repository: %v", err)
	}

	// 3. Initialize Services
	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	bookService := service.NewBookService(bookRepo)
	orderService := service.NewOrderService(orderRepo, bookRepo)

	// 4. Initialize Handlers & Middleware
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	bookHandler := handler.NewBookHandler(bookService)
	orderHandler := handler.NewOrderHandler(orderService)
	authMW := middleware.NewAuthMiddleware(jwtSecret)

	// 5. Setup Router
	r := router.SetupRoutes(router.RouterConfig{
		AuthHandler:  authHandler,
		UserHandler:  userHandler,
		BookHandler:  bookHandler,
		OrderHandler: orderHandler,
		AuthMW:       authMW,
	})

	// 6. Wrap router in global Middleware Pipeline (Logging, Panic Recovery, CORS)
	httpHandler := middleware.Chain(
		r,
		middleware.LoggingMiddleware,
		middleware.RecoveryMiddleware,
		middleware.CORSMiddleware,
	)

	// 7. Configure and Start HTTP Server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("Server running on http://localhost:8080")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server stopped unexpectedly: %v", err)
	}
}