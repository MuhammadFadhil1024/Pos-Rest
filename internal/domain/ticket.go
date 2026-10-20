package domain

import (
	"context"
	"time"
)

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
	MenuId   int64   `json:"menu_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Subtotal float64 `json:"subtotal"`
}

type TicketFilter struct {
	BranchID    int64
	Search      string
	Status      string
	ServiceID   int64
	TotalAmount float64
	StartDate   time.Time
	EndDate     time.Time
	Limit       int
	Offset      int
}

type TicketListItem struct {
	ID           int64     `json:"id"`
	TicketNumber string    `json:"ticket_number"`
	BranchId     int64     `json:"branch_id"`
	UserId       int64     `json:"user_id"`
	CashierName  string    `json:"cashier_name,omitempty"`
	ServiceId    int64     `json:"service_id"`
	ServiceName  string    `json:"service_name,omitempty"`
	TotalAmount  float64   `json:"total_amount"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type TicketRepository interface {
	CreateFullTicketWithTx(ctx context.Context, ticket *Ticket, items []TicketItemInsertDb, payments []TicketPaymentInsertDb) error
	GetRecentTickets(ctx context.Context, filter TicketFilter) ([]*TicketListItem, int64, error)
}

type TicketService interface {
	CreateTicket(ctx context.Context, branchId int64, serviceId int64, userId int64, items []TicketItemInput, payments []TicketPaymentInput) (*Ticket, error)
	GetRecentTickets(ctx context.Context, filter TicketFilter) ([]*TicketListItem, int64, error)
}
