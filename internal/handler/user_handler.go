package handler

import (
	"encoding/json"
	"net/http"

	"bookstore-api/internal/dto"
	"bookstore-api/internal/middleware"
	"bookstore-api/internal/service"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterUserRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	response, err := h.userService.Register(r.Context(), req)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusCreated, response)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeJSONError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	response, err := h.userService.GetUserByID(r.Context(), userID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, response)
}