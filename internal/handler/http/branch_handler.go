package handler

import (
	"encoding/json"
	"net/http"
	"pos-backend/internal/domain"
	"pos-backend/internal/handler/http/middleware"
	"pos-backend/pkg/jwtutil"
)

type branchHandler struct {
	branchService domain.BranchService
}

func NewBranchHandler (service domain.BranchService) *branchHandler {
	return &branchHandler{branchService: service}
}

func (h *branchHandler) GeMyBranches(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	claimsVal := r.Context().Value(middleware.UserContextKey)

	if claimsVal == nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	claims, ok := claimsVal.(*jwtutil.CustomClaims)

	if !ok  {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	branches, err := h.branchService.GetUserBranches(r.Context(), claims.UserID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   branches,
	})
}