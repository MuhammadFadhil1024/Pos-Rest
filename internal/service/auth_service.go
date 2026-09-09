package service

import (
	"context"
	"errors"
	"pos-backend/internal/domain"
	"pos-backend/pkg/jwtutil"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo domain.UserRepository
	jwtSecret string
}

func NewAuthService(repo domain.UserRepository, secret string) domain.AuthService {
	return &authService{userRepo: repo, jwtSecret: secret}
}

func (s *authService) Register(ctx context.Context, email, role, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return  nil, err
	}

	user := &domain.User{
		Email: email,
		Password: string(hashedPassword),
		Role: role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (string, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", errors.New("wrong email or pasword!")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", errors.New("wrong email or pasword!")
	}

	token, err := jwtutil.GenerateToken(user.ID, user.Role, s.jwtSecret, 24*time.Hour)

	if err != nil {
		return "", err
	}

	return token, nil
}