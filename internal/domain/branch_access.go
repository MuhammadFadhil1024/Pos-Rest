package domain

import "context"

// BranchAccessRepository checks whether a user is allowed to access a branch.
type BranchAccessRepository interface {
	UserHasBranchAccess(ctx context.Context, userID, branchID int64) (bool, error)
}
