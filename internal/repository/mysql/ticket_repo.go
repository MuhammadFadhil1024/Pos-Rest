package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"pos-backend/internal/domain"
	"strings"
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

func (r *ticketRepository) GetRecentTickets(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error) {
	var conditions []string
	var args []interface{}

	// Branch ID filter (required)
	conditions = append(conditions, "t.branch_id = ?")
	args = append(args, filter.BranchID)

	// Date range: 3 days window
	if !filter.StartDate.IsZero() {
		conditions = append(conditions, "t.created_at >= ?")
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		conditions = append(conditions, "t.created_at < ?")
		args = append(args, filter.EndDate)
	}

	log.Printf("Filter conditions: %v, args: %v", conditions, args)

	// Optional search (partial ticket number)
	if filter.Search != "" {
		conditions = append(conditions, "t.ticket_number LIKE ?")
		args = append(args, "%"+filter.Search+"%")
	}

	// Optional status
	if filter.Status != "" {
		conditions = append(conditions, "t.status = ?")
		args = append(args, filter.Status)
	}

	// Optional service_id
	if filter.ServiceID > 0 {
		conditions = append(conditions, "t.service_id = ?")
		args = append(args, filter.ServiceID)
	}

	// Optional total_amount
	if filter.TotalAmount > 0 {
		conditions = append(conditions, "t.total_amount = ?")
		args = append(args, filter.TotalAmount)
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total items
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM ticket t WHERE %s", whereClause)
	var totalItems int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalItems)
	if err != nil {
		return nil, 0, err
	}

	if totalItems == 0 {
		return []*domain.TicketListItem{}, 0, nil
	}

	// Select items
	selectQuery := fmt.Sprintf(`
		SELECT 
			t.id, t.branch_id, t.user_id, COALESCE(u.email, '') as cashier_name,
			t.service_id, COALESCE(s.name, '') as service_name,
			t.ticket_number, t.total_amount, t.status, t.created_at
		FROM ticket t
		LEFT JOIN users u ON t.user_id = u.id
		LEFT JOIN services s ON t.service_id = s.id
		WHERE %s
		ORDER BY t.created_at DESC, t.id DESC
		LIMIT ? OFFSET ?
	`, whereClause)

	selectArgs := make([]interface{}, len(args), len(args)+2)
	copy(selectArgs, args)
	selectArgs = append(selectArgs, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, selectQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []*domain.TicketListItem
	for rows.Next() {
		var item domain.TicketListItem
		if err := rows.Scan(
			&item.ID,
			&item.BranchId,
			&item.UserId,
			&item.CashierName,
			&item.ServiceId,
			&item.ServiceName,
			&item.TicketNumber,
			&item.TotalAmount,
			&item.Status,
			&item.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		tickets = append(tickets, &item)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, err
	}

	return tickets, totalItems, nil
}
