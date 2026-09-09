package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"pos-backend/internal/domain"
	"strconv"
)

type preloadHandler struct {
	serviceTypeService 	domain.ServiceTypeService
	branchService		domain.BranchService
	paymentService		domain.PaymentService
}

func NewPreloadHandler (service domain.ServiceTypeService, branch domain.BranchService, payment domain.PaymentService) *preloadHandler {
	return &preloadHandler{
		serviceTypeService: service,
		branchService: branch,
		paymentService: payment,
	}
}

func (h *preloadHandler) GetPreloadData(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	idBranch := r.PathValue("id")

	
	id, err := strconv.ParseInt(idBranch, 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string]string{"error": "ID not valid"})
        return
	}

	services, err := h.serviceTypeService.GetBranchServices(r.Context(), id)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	branch, err := h.branchService.GetUserBranch(r.Context(), id)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	menus, err := h.serviceTypeService.GetBranchServiceMenu(r.Context(), id)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	

	response := map[string]interface{}{
		"status": "success",
		"data" : map[string]interface{}{
			"branch": branch,
			"services": services,
			"menus": menus,
		},
	}

	json.NewEncoder(w).Encode(response)
}