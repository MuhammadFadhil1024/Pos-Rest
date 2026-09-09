package domain

type TicketItem struct {
	ID       int64   `json:"id"`
	TicketId int64   `json:"ticket_id"`
	MenuId   int64   `json:"menu_id"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Subtotal float64 `json:"subtotal"`
}

type TicketItemRepository interface {
	CreateTicketItem(ticketItem *TicketItem) error
}

type TicketItemService interface {
	CreateTicketItem(ticketItem *TicketItem) error
}