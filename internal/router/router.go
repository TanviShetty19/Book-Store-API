package router

import (
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
)

// NewRouter constructs and configures the core HTTP ServeMux routing table
func NewRouter(bookHandler *handler.BookHandler, authHandler *handler.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Public Auth Endpoints
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Public Book Read Endpoints
	mux.HandleFunc("GET /books", bookHandler.GetAllBooks)
	mux.HandleFunc("GET /books/{id}", bookHandler.GetBookByID)

	// Protected Book Write Endpoints (Requires JWT)
	mux.HandleFunc("POST /books", middleware.AuthMiddleware(bookHandler.CreateBook))
	mux.HandleFunc("PUT /books/{id}", middleware.AuthMiddleware(bookHandler.UpdateBook))

	// Admin-Only Book Delete Endpoint (Requires JWT + Admin Role)
	mux.HandleFunc("DELETE /books/{id}", middleware.RequireRole("admin", bookHandler.DeleteBook))

	return mux
}