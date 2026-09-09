package mysql

import (
	"context"
	"database/sql"
	"pos-backend/internal/domain"
)

type ticketRepository struct {
	db *sql.DB
}

func NewTicketRepository(db *sql.DB) domain.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) CreateFullTicketWithTx(ctx context.Context, ticket *domain.Ticket, items []domain.TicketItemInsertDb, payments []domain.TicketPaymentInsertDb) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. Insert Ticket
	queryTicket := `INSERT INTO ticket (branch_id, user_id, service_id, ticket_number, total_amount, status) 
	VALUES (?, ?, ?, ?, ?, ?)`

	res, err := tx.ExecContext(ctx, queryTicket, ticket.BranchId, ticket.UserId, ticket.ServiceId, ticket.TicketNumber, ticket.TotalAmount, ticket.Status)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	ticket.ID = id

	// 2. Insert Ticket Items
	queryItem := `INSERT INTO ticket_item (ticket_id, menu_id, quantity, price, subtotal) VALUES (?, ?, ?, ?, ?)`
	for _, item := range items {
		_, err = tx.ExecContext(ctx, queryItem, ticket.ID, item.MenuId, item.Quantity, item.Price, item.Subtotal)
		if err != nil {
			return err
		}
	}

	// 3. Insert Ticket Payments
	queryPayment := `INSERT INTO ticket_payment (ticket_id, payment_type_id, amount, change_amount) VALUES (?, ?, ?, ?)`
	for _, payment := range payments {
		_, err = tx.ExecContext(ctx, queryPayment, ticket.ID, payment.PaymentTypeId, payment.Amount, payment.ChangeAmount)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}
