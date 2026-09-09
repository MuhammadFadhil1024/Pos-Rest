package service

import (
	"context"
	"pos-backend/internal/domain"
)

type branchService struct {
	branchRepo domain.BranchRepository
}

func NewBranchService(repo domain.BranchRepository) domain.BranchService {
	return &branchService{branchRepo: repo}
}

func (s *branchService) GetUserBranches(ctx context.Context, id int64) ([]*domain.Branch, error) {
	return s.branchRepo.GetByUserId(ctx, id)
}

func (s *branchService) GetUserBranch(ctx context.Context, branchId int64) (*domain.Branch, error) {
	return  s.branchRepo.GetByBranchId(ctx, branchId)
}