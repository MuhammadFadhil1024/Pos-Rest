package domain

import "context"

type Service struct {
	ID       int64      `json:"id"`
	Name     string     `json:"service_name"`
	Icon     string     `json:"icon"`
	Payments []*Payment `json:"payments,omitempty"`
}

type ServiceRepository interface {
	GetByBranchId(ctx context.Context, id int64) ([]*Service, error)
	GetServiceMenu(ctx context.Context, branch_id int64) ([]*ServiceMenu, error)
}

type ServiceTypeService interface {
	GetBranchServices(ctx context.Context, id int64) ([]*Service, error)
	GetBranchServiceMenu(ctx context.Context, branch_id int64) ([]*ServiceMenu, error)
}