package handler

import (
	"encoding/json"
	"net/http"
	"partying/internal/service"
	"partying/pkg/response"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.authService.Register(input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.Created(w, result)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	result, err := h.authService.Login(input)
	if err != nil {
		response.BadRequest(w, err.Error())
		return
	}

	response.OK(w, result)
}
