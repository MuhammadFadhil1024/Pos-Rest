package domain

import "context"

type User struct {
	ID int64 `json:"id"`
	Email string `json:"email"`
	Password string `json:"-"` // tidak kembalikan password
	Role string `json:"role"`
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error) 
}

type AuthService interface {
	Register( ctx context.Context, email, password, role string) (*User, error)
	Login(ctx context.Context, email, password string) (string, error) // Return JWT Token
}