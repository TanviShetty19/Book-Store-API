package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"bookstore-api/internal/db"
	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/router"
	"bookstore-api/internal/service"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// 1. Load environment variables from .env file (if present)
	if err := godotenv.Load(); err != nil {
		log.Println("Note: No .env file found; reading configuration directly from system environment")
	}

	// 2. Read Configuration from Environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required; server refusing to start without it")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8080" // Default HTTP port
	}

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "mongo" // Default storage strategy
	}

	var bookRepo repository.BookRepository
	var userRepo repository.UserRepository
	var orderRepo repository.OrderRepository
	var mongoClient *mongo.Client

	// 3. Initialize Repositories based on STORAGE_TYPE
	switch storageType {
	case "mongo":
		mongoURI := os.Getenv("MONGO_URI")
		if mongoURI == "" {
			log.Fatal("MONGO_URI environment variable is required when STORAGE_TYPE=mongo")
		}

		dbName := os.Getenv("MONGO_DB_NAME")
		if dbName == "" {
			log.Fatal("MONGO_DB_NAME environment variable is required when STORAGE_TYPE=mongo")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		client, mongoDB, err := db.ConnectMongo(ctx, mongoURI, dbName)
		if err != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", err)
		}
		mongoClient = client

		log.Println("Initializing MongoDB repositories...")
		bookRepo = repository.NewMongoBookRepository(mongoDB)
		userRepo = repository.NewMongoUserRepository(mongoDB)
		orderRepo = repository.NewMongoOrderRepository(mongoDB, bookRepo)

	case "json":
		log.Println("Initializing JSON file repositories...")
		jsonBookRepo, err := repository.NewJsonBookRepository("data/books.json")
		if err != nil {
			log.Fatalf("Failed to initialize book repository: %v", err)
		}

		jsonUserRepo, err := repository.NewJsonUserRepository("data/users.json")
		if err != nil {
			log.Fatalf("Failed to initialize user repository: %v", err)
		}

		jsonOrderRepo, err := repository.NewJsonOrderRepository("data/orders.json", jsonBookRepo)
		if err != nil {
			log.Fatalf("Failed to initialize order repository: %v", err)
		}

		bookRepo = jsonBookRepo
		userRepo = jsonUserRepo
		orderRepo = jsonOrderRepo

	default:
		log.Fatalf("Invalid STORAGE_TYPE '%s'. Must be 'mongo' or 'json'", storageType)
	}

	// Ensure MongoDB connection is gracefully closed when the app shuts down
	defer func() {
		if mongoClient != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := mongoClient.Disconnect(ctx); err != nil {
				log.Printf("Error disconnecting from MongoDB: %v", err)
			}
		}
	}()

	// 4. Initialize Services
	authService := service.NewAuthService(userRepo, jwtSecret)
	userService := service.NewUserService(userRepo)
	bookService := service.NewBookService(bookRepo)
	orderService := service.NewOrderService(orderRepo, bookRepo)

	// 5. Initialize Handlers & Middleware
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	bookHandler := handler.NewBookHandler(bookService)
	orderHandler := handler.NewOrderHandler(orderService)
	authMW := middleware.NewAuthMiddleware(jwtSecret)

	// 6. Setup Router
	r := router.SetupRoutes(router.RouterConfig{
		AuthHandler:  authHandler,
		UserHandler:  userHandler,
		BookHandler:  bookHandler,
		OrderHandler: orderHandler,
		AuthMW:       authMW,
	})

	// 7. Wrap router in global Middleware Pipeline
	httpHandler := middleware.Chain(
		r,
		middleware.LoggingMiddleware,
		middleware.RecoveryMiddleware,
		middleware.CORSMiddleware,
	)

	// 8. Configure and Start HTTP Server with Graceful Shutdown
	server := &http.Server{
		Addr:         port,
		Handler:      httpHandler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server running on http://localhost%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server stopped unexpectedly: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server gracefully...")

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelShutdown()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly")
}