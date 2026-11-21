package mysql

import (
	"context"
	"pos-backend/internal/domain"
)

type userBranchesRepository struct {
	db DB
}

func NewUserBranchesRepository(db DB) domain.UserBranchesRepository {
	return &userBranchesRepository{db: db}
}

func (r *userBranchesRepository) CreateUserBranches(userId int64, branchId int64) (*domain.UserBranches, error) {
	query := `INSERT INTO mm_user_branch (user_id, branch_id) VALUES (?, ?)`

	_, err := r.db.ExecContext(context.Background(), query, userId, branchId)
	if err != nil {
		return nil, err
	}

	return nil, nil
}