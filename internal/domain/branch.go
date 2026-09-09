package domain

import "context"

type Branch struct {
	ID int64 `json:"branch_id"`
	Name string `json:"name"`
}

type BranchRepository interface {
	GetByUserId(ctx context.Context, id int64) ([]*Branch, error)
	GetByBranchId(ctx context.Context, branchId int64) (*Branch, error)
}

type BranchService interface {
	GetUserBranches(ctx context.Context, id int64) ([]*Branch, error)
	GetUserBranch(ctx context.Context, branchId int64) (*Branch, error)
}