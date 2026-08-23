package handler

import (
	"encoding/json"
	"net/http"

	"bookstore-api/internal/dto"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/service"

	"github.com/gorilla/mux"
)

type OrderHandler struct {
	orderService service.OrderService
}

func NewOrderHandler(orderService service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized: missing user context")
		return
	}

	var req dto.CreateOrderRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.orderService.CreateOrder(r.Context(), userID, req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)

	response, err := h.orderService.GetOrderByID(r.Context(), userID, userRole, orderID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to fetch orders")
		return
	}

	writeJSONResponse(w, http.StatusOK, orders)
}

// Shared JSON response helpers
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}