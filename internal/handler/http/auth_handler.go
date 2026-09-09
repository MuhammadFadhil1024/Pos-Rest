package handler

import (
	"encoding/json"
	"net/http"
	"pos-backend/internal/domain"
)

type authHandler struct {
	authService domain.AuthService
}

func NewAuthHandler (service domain.AuthService) *authHandler {
	return &authHandler{authService: service}
}

func (h *authHandler) Login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
			Email		string `json:"email"`
			Password	string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{
				"error":"Invalid status",
			})
			return
		}

		token, err := h.authService.Login(r.Context(), req.Email, req.Password)

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized) // Status 401 Unauthorized
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":"success",
			"message": "Success Login",
			"token" : token,
		})
}