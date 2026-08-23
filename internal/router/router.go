package router

import (
	"net/http"

	"bookstore-api/internal/handler"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/model"

	"github.com/gorilla/mux"
)

type RouterConfig struct {
	AuthHandler  *handler.AuthHandler
	UserHandler  *handler.UserHandler
	BookHandler  *handler.BookHandler
	OrderHandler *handler.OrderHandler
	AuthMW       *middleware.AuthMiddleware
}

func SetupRoutes(cfg RouterConfig) *mux.Router {
	r := mux.NewRouter()

	// --- Public Auth Routes ---
	r.HandleFunc("/auth/login", cfg.AuthHandler.Login).Methods(http.MethodPost)
	r.HandleFunc("/users/register", cfg.UserHandler.Register).Methods(http.MethodPost)

	// --- Public Book Browsing ---
	r.HandleFunc("/books", cfg.BookHandler.GetAllBooks).Methods(http.MethodGet)
	r.HandleFunc("/books/{id}", cfg.BookHandler.GetBookByID).Methods(http.MethodGet)

	// --- Protected Routes (Requires JWT Authentication) ---
	api := r.PathPrefix("").Subrouter()
	api.Use(cfg.AuthMW.Authenticate)

	// User Profile
	api.HandleFunc("/users/me", cfg.UserHandler.GetProfile).Methods(http.MethodGet)

	// Orders (Customer & Admin)
	api.HandleFunc("/orders", cfg.OrderHandler.CreateOrder).Methods(http.MethodPost)
	api.HandleFunc("/orders", cfg.OrderHandler.GetMyOrders).Methods(http.MethodGet)
	api.HandleFunc("/orders/{id}", cfg.OrderHandler.GetOrderByID).Methods(http.MethodGet)

	// --- Admin-Only Book Operations ---
	adminBooks := api.PathPrefix("/books").Subrouter()
	// Matches model.RoleAdmin ("ADMIN")
	adminBooks.Use(middleware.RequireRole(string(model.RoleAdmin)))

	adminBooks.HandleFunc("", cfg.BookHandler.CreateBook).Methods(http.MethodPost)
	adminBooks.HandleFunc("/{id}", cfg.BookHandler.UpdateBook).Methods(http.MethodPut, http.MethodPatch)
	adminBooks.HandleFunc("/{id}", cfg.BookHandler.DeleteBook).Methods(http.MethodDelete)

	return r
}