package main

import (
	"fmt"
	"log"
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/router"
	"bookstore-api/internal/service"
)

func main() {
	// 1. Initialize Sam (Repository / Warehouse)
	repo := repository.NewJSONBookRepository("data/books.json")

	// 2. Initialize Bob (Service / Business Rules)
	svc := service.NewBookService(repo)

	// 3. Initialize Carl (Handler / Receptionist)
	h := handler.NewBookHandler(svc)

	// 4. Initialize Router (Switchboard)
	r := router.NewRouter(h)

	// 5. Start Server
	fmt.Println("Bookstore API server running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}