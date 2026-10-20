package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"pos-backend/internal/domain"
	"strconv"
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

func (h *ticketHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	branchIDStr := r.URL.Query().Get("branch_id")
	branchID, err := strconv.ParseInt(branchIDStr, 10, 64)
	if err != nil || branchID <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid or missing branch_id"})
		return
	}

	search := r.URL.Query().Get("search")
	status := r.URL.Query().Get("status")

	var serviceID int64
	if serviceIDStr := r.URL.Query().Get("service_id"); serviceIDStr != "" {
		if sid, err := strconv.ParseInt(serviceIDStr, 10, 64); err == nil {
			serviceID = sid
		}
	}

	var totalAmount float64
	if totalAmountStr := r.URL.Query().Get("total_amount"); totalAmountStr != "" {
		if ta, err := strconv.ParseFloat(totalAmountStr, 64); err == nil {
			totalAmount = ta
		}
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 20
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	filter := domain.TicketFilter{
		BranchID:    branchID,
		Search:      search,
		Status:      status,
		ServiceID:   serviceID,
		TotalAmount: totalAmount,
		Limit:       limit,
		Offset:      offset,
	}

	tickets, totalItems, err := h.ticketService.GetRecentTickets(r.Context(), filter)
	if err != nil {
		log.Printf("Failed to get tickets: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to fetch tickets"})
		return
	}

	if tickets == nil {
		tickets = []*domain.TicketListItem{}
	}

	totalPages := 0
	if totalItems > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(limit)))
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "success",
		"data":   tickets,
		"meta": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_items": totalItems,
			"total_pages": totalPages,
		},
	})
}
