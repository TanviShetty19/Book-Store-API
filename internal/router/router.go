package router

import (
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
)

func NewRouter(bookHandler *handler.BookHandler, authHandler *handler.AuthHandler) *http.ServeMux {
	mux := http.NewServeMux()

	// Public Auth Endpoints
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)

	// Public Read Endpoints
	mux.HandleFunc("GET /books", bookHandler.GetAllBooks)
	mux.HandleFunc("GET /books/{id}", bookHandler.GetBookByID)

	// Protected Batch Endpoints
	mux.HandleFunc("POST /books/batch", middleware.AuthMiddleware(bookHandler.CreateBatchBooks))
	mux.HandleFunc("DELETE /books/batch", middleware.RequireRole("admin", bookHandler.DeleteBatchBooks))

	// Protected Single Book Write Endpoints (Requires JWT)
	mux.HandleFunc("POST /books", middleware.AuthMiddleware(bookHandler.CreateBook))
	mux.HandleFunc("PUT /books/{id}", middleware.AuthMiddleware(bookHandler.UpdateBook))

	// Admin-Only Single Book Delete Endpoint (Requires JWT + Admin Role)
	mux.HandleFunc("DELETE /books/{id}", middleware.RequireRole("admin", bookHandler.DeleteBook))

	return mux
}