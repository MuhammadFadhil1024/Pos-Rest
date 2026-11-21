package domain

import "context"

type User struct {
	ID          int64  `json:"id"`
	Email       string `json:"email"`
	Password    string `json:"-"` // tidak kembalikan password
	Role        string `json:"role"`
	PhoneNumber string `json:"phone_number"`
}

type UserInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	PhoneNumber string `json:"phone_number"`
}

type RegistrationResult struct {
	User   *User   `json:"user"`
	Branch *Branch `json:"branch"`
}

type RegistrationRepositories struct {
	UserRepo         UserRepository
	BranchRepo       BranchRepository
	UserBranchesRepo UserBranchesRepository
}

type RegistrationUnitOfWork interface {
	WithinTransaction(
		ctx context.Context,
		fn func(repos RegistrationRepositories) error,
	) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
}

type AuthService interface {
	Register(ctx context.Context, branchInputs []BranchInput, userInputs []UserInput) (*RegistrationResult, error)
	Login(ctx context.Context, email, password string) (string, error) // Return JWT Token
}
