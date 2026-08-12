package main

import (
	"fmt"
	"log"
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/repository"
	"bookstore-api/internal/service"
)

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
	//Start Server
	fmt.Println("Bookstore API server running on http://localhost:8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}