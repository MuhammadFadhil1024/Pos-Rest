package domain

type TicketPayment struct {
	ID            int64   `json:"id"`
	TicketId      int64   `json:"ticket_id"`
	PaymentTypeId int64   `json:"payment_type_id"`
	Amount        float64 `json:"amount"`
	ChangeAmount  float64 `json:"change_amount"`
}

type TicketPaymentInput struct {
	PaymentTypeId int64   `json:"payment_type_id"`
	Amount        float64 `json:"amount"`
}

type TicketPaymentInsertDb struct {
	PaymentTypeId int64   `json:"payment_type_id"`
	Amount        float64 `json:"amount"`
	ChangeAmount  float64 `json:"change_amount"`
}

type TicketPaymentRepository interface {
	CreateTicketPayment(ticketPayment *TicketPayment) error
}

type TicketPaymentService interface {
	CreateTicketPayment(ticketPayment *TicketPayment) error
}
