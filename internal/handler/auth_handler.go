package handler

import (
	"encoding/json"
	"net/http"

	"bookstore-api/internal/dto"
	"bookstore-api/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequestDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid request payload")
		return
	}

	token, err := h.authService.Login(r.Context(), req)
	if err != nil {
		writeJSONError(w, mapErrorToStatusCode(err), err.Error())
		return
	}

	writeJSONResponse(w, http.StatusOK, dto.AuthResponseDTO{
		Token:     token,
		TokenType: "Bearer",
	})
}