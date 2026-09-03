package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"bookstore-api/internal/apperrors"
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
		// Dynamically maps ErrConflict -> 409, ErrNotFound -> 404, etc.
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

func (h *OrderHandler) ConfirmOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized: missing user context")
		return
	}
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)

	response, err := h.orderService.ConfirmOrder(r.Context(), userID, userRole, orderID)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrderHandler) CancelOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized: missing user context")
		return
	}
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)

	response, err := h.orderService.CancelOrder(r.Context(), userID, userRole, orderID)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}

func (h *OrderHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderID := vars["id"]

	userID, _ := r.Context().Value(middleware.UserIDKey).(string)
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(string)

	response, err := h.orderService.GetOrderByID(r.Context(), userID, userRole, orderID)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
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

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	var offset, limit int64
	if offsetStr != "" {
		offset, _ = strconv.ParseInt(offsetStr, 10, 64)
	}
	if limitStr != "" {
		limit, _ = strconv.ParseInt(limitStr, 10, 64)
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID, offset, limit)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, orders)
}

// Helper to inspect apperrors sentinel values and select proper REST status codes
func mapErrorToStatusCode(err error) int {
	switch {
	case errors.Is(err, apperrors.ErrNotFound):
		return http.StatusNotFound // 404
	case errors.Is(err, apperrors.ErrConflict):
		return http.StatusConflict // 409
	case errors.Is(err, apperrors.ErrForbidden):
		return http.StatusForbidden // 403
	case errors.Is(err, apperrors.ErrUnauthorized):
		return http.StatusUnauthorized // 401
	case errors.Is(err, apperrors.ErrValidation):
		return http.StatusBadRequest // 400
	default:
		return http.StatusInternalServerError // 500
	}
}

// Shared JSON response helpers
func writeJSONError(w http.ResponseWriter, statusCode int, message string) {
	writeJSONResponse(w, statusCode, map[string]string{"error": message})
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}
