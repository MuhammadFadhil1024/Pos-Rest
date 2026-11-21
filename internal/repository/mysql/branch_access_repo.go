package mysql

import (
	"context"
	"pos-backend/internal/domain"
)

type branchAccessRepository struct {
	db DB
}

func NewBranchAccessRepository(db DB) domain.BranchAccessRepository {
	return &branchAccessRepository{db: db}
}

func (r *branchAccessRepository) UserHasBranchAccess(ctx context.Context, userID, branchID int64) (bool, error) {
	const query = `SELECT EXISTS(
		SELECT 1 FROM mm_user_branch WHERE user_id = ? AND branch_id = ?
	)`

	var exists bool
	if err := r.db.QueryRowContext(ctx, query, userID, branchID).Scan(&exists); err != nil {
		return false, err
	}

	return exists, nil
}
