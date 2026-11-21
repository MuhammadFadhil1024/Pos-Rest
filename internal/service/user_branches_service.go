package service

import "pos-backend/internal/domain"

type userBranchesService struct {
	userBranchesRepo domain.UserBranchesRepository
}

func NewUserBranchesService(repo domain.UserBranchesRepository) domain.UserBranchesService {
	return &userBranchesService{userBranchesRepo: repo}
}

func (s *userBranchesService) CreateUserBranches(userId int64, branchId int64) (*domain.UserBranches, error) {
	return s.userBranchesRepo.CreateUserBranches(userId, branchId)
}