package handler

import (
	"encoding/json"
	"net/http"
	"pos-backend/internal/domain"
	"strings"
)

type registrationHandler struct {
	branchService domain.BranchService
	userService   domain.AuthService
}

func NewRegistrationHandler(branchService domain.BranchService, userService domain.AuthService) *registrationHandler {
	return &registrationHandler{
		branchService: branchService,
		userService:   userService,
	}
}

func (h *registrationHandler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Name        string `json:"name"`
		CompanyName string `json:"company_name"`
		Email       string `json:"email"`
		Password    string `json:"password"`
		PhoneNumber string `json:"phone_number"`
		Address     string `json:"address"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Invalid request payload",
		})
		return
	}

	defer r.Body.Close()

	// All request fields are required.
	var missingFields []string
	if strings.TrimSpace(req.Name) == "" {
		missingFields = append(missingFields, "name")
	}
	if strings.TrimSpace(req.CompanyName) == "" {
		missingFields = append(missingFields, "company_name")
	}
	if strings.TrimSpace(req.Email) == "" {
		missingFields = append(missingFields, "email")
	}
	if strings.TrimSpace(req.Password) == "" {
		missingFields = append(missingFields, "password")
	}
	if strings.TrimSpace(req.PhoneNumber) == "" {
		missingFields = append(missingFields, "phone_number")
	}
	if strings.TrimSpace(req.Address) == "" {
		missingFields = append(missingFields, "address")
	}

	if len(missingFields) > 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "missing required fields: " + strings.Join(missingFields, ", "),
		})
		return
	}

	var branchInputs []domain.BranchInput
	branchInputs = append(branchInputs, domain.BranchInput{
		Name:        req.Name,
		CompanyName: req.CompanyName,
		Address:     req.Address,
	})

	var userInputs []domain.UserInput
	userInputs = append(userInputs, domain.UserInput{
		Email:       req.Email,
		Password:    req.Password,
		PhoneNumber: req.PhoneNumber,
	})

	user, err := h.userService.Register(r.Context(), branchInputs, userInputs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "error",
			"message": "Failed to register user. " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "Success Register",
		"data":    user,
	})
}
