package domain

import "context"

type Ticket struct {
	ID           int64   `json:"id"`
	BranchId     int64   `json:"branch_id"`
	UserId       int64   `json:"user_id"`
	ServiceId    int64   `json:"service_id"`
	TicketNumber string  `json:"ticket_number"`
	TotalAmount  float64 `json:"total_amount"`
	Status       string  `json:"status"`
}

type TicketItemInput struct {
	MenuId   int64 `json:"menu_id"`
	Quantity int   `json:"quantity"`
}

type TicketItemInsertDb struct {
	MenuId   int64 `json:"menu_id"`
	Quantity int   `json:"quantity"`
	Price	float64 `json:"price"`
	Subtotal float64 `json:"subtotal"`
}

type TicketRepository interface {
	CreateFullTicketWithTx(ctx context.Context, ticket *Ticket, items []TicketItemInsertDb, payments []TicketPaymentInsertDb) error
}

type TicketService interface {
	CreateTicket(ctx context.Context, branchId int64, serviceId int64, userId int64, items []TicketItemInput, payments []TicketPaymentInput) (*Ticket, error)
}
