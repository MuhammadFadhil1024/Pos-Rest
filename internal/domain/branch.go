package domain

import "context"

type Branch struct {
	ID          int64  `json:"branch_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	CompanyName string `json:"company_name"`
	Address     string `json:"address"`
}

type BranchInput struct {
	Name        string `json:"name"`
	CompanyName string `json:"company_name"`
	Address     string `json:"address"`
}

type BranchRepository interface {
	GetByUserId(ctx context.Context, id int64) ([]*Branch, error)
	GetByBranchId(ctx context.Context, branchId int64) (*Branch, error)
	CreateBranch(ctx context.Context, branch *Branch) error
}

type BranchService interface {
	GetUserBranches(ctx context.Context, id int64) ([]*Branch, error)
	GetUserBranch(ctx context.Context, branchId int64) (*Branch, error)
}
