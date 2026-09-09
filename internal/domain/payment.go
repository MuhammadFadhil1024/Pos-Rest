package domain

import "context"

type Payment struct {
	ID 			int64 `json:"id"`
	ServiceID 	int64 `json:"service_id"`
	BranchID 	int64 `json:"branch_id"`
	Name 		string `json:"name"`
}

type PaymentRepository interface {
	GetByService(ctx context.Context, id int64) ([]*Payment, error)
}

type PaymentService interface {
	GetServicePayments(ctx context.Context, id int64) ([]*Payment, error)
}