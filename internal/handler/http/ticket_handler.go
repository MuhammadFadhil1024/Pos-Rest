package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"pos-backend/internal/domain"
)

type ticketHandler struct {
	ticketService domain.TicketService
}

func NewTicketHandler(ticketService domain.TicketService) *ticketHandler {
	return &ticketHandler{
		ticketService: ticketService,
	}
}

type CreateTicketRequest struct {
	BranchId int64 `json:"branch_id"`
	ServiceId int64 `json:"service_id"`
	UserId   int64 `json:"user_id"`
	Items    []struct {
		MenuId   int64 `json:"menu_id"`
		Quantity int   `json:"quantity"`
	} `json:"items"`
	Payments []struct {
		PaymentTypeId int64   `json:"payment_type_id"`
		Amount        float64 `json:"amount"`
	} `json:"payments"`
}

func (h *ticketHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	var req CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var itemInputs []domain.TicketItemInput
	for _, item := range req.Items {
		itemInputs = append(itemInputs, domain.TicketItemInput{
			MenuId:   item.MenuId,
			Quantity: item.Quantity,
		})
	}

	var paymentInputs []domain.TicketPaymentInput
	for _, p := range req.Payments {
		paymentInputs = append(paymentInputs, domain.TicketPaymentInput{
			PaymentTypeId: p.PaymentTypeId,
			Amount:        p.Amount,
		})
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), req.BranchId, req.ServiceId, req.UserId, itemInputs, paymentInputs)
	if err != nil {
		log.Printf("Failed to create ticket: %v", err)
		http.Error(w, "Failed to create ticket", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   ticket,
	})
}
