package domain

type UserBranches struct {
	UserId   int64 `json:"user_id"`
	BranchId int64 `json:"branch_id"`
}

type UserBranchesService interface {
	CreateUserBranches(userId int64, branchId int64) (*UserBranches, error)
}

type UserBranchesRepository interface {
	CreateUserBranches(userId int64, branchId int64) (*UserBranches, error)
}