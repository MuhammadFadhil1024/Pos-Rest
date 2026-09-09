package mysql

import (
	"context"
	"database/sql"
	"pos-backend/internal/domain"
	"strings"
)

type paymentRepository struct {
	db *sql.DB
	branchServiceRepo domain.ServiceRepository
}

func NewPaymentRepository(db *sql.DB, branchServiceRepo domain.ServiceRepository) *paymentRepository {
	return &paymentRepository{db: db, branchServiceRepo: branchServiceRepo}
}

func (r *paymentRepository) GetByService(ctx context.Context, id int64) ([]*domain.Payment, error) {
	// get active services by branch id
	services, err := r.branchServiceRepo.GetByBranchId(ctx, id)
	if err != nil {
		return nil, err
	}

	if len(services) == 0 {
		return []*domain.Payment{}, nil
	}

	// get service ids
	var serviceIds []interface{}
	for _, service := range services {
		serviceIds = append(serviceIds, service.ID)
	}

	query := `SELECT s.id as service_id, p.id as payment_id, p.name as payment_name
		FROM services s
		JOIN mm_service_payment sp ON s.id = sp.service_id
		JOIN payment_type p ON sp.payment_type_id = p.id
		WHERE s.is_active = 1 AND p.is_active = 1 AND s.id IN (?` + strings.Repeat(",?", len(serviceIds)-1) + `)
	`

	rows, err := r.db.QueryContext(ctx, query, serviceIds...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ServiceID, &p.ID, &p.Name); err != nil {
			return nil, err
		}
		payments = append(payments, &p)
	}

	return payments, nil
}
