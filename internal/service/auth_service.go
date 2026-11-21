package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"pos-backend/internal/domain"
	"pos-backend/pkg/jwtutil"
	"pos-backend/pkg/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	userRepo   domain.UserRepository
	jwtSecret  string
	unitOfWork domain.RegistrationUnitOfWork
}

func NewAuthService(repo domain.UserRepository, secret string, unitOfWork domain.RegistrationUnitOfWork) domain.AuthService {
	return &authService{userRepo: repo, jwtSecret: secret, unitOfWork: unitOfWork}
}

func (s *authService) Register(ctx context.Context, branchInputs []domain.BranchInput, userInputs []domain.UserInput) (*domain.RegistrationResult, error) {
	if len(userInputs) == 0 {
		return nil, errors.New("no user inputs provided")
	}
	if len(branchInputs) == 0 {
		return nil, errors.New("no branch inputs provided")
	}

	userInput := userInputs[0]
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(userInput.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Email:       userInput.Email,
		Password:    string(hashedPassword),
		Role:        "admin", // Set default role to "admin"
		PhoneNumber: userInput.PhoneNumber,
	}

	branch := &domain.Branch{
		Name:        branchInputs[0].Name,
		Slug:        utils.GenerateSlug(branchInputs[0].Name),
		CompanyName: branchInputs[0].CompanyName,
		Address:     branchInputs[0].Address,
	}

	//chek if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, user.Email)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	if existingUser != nil {
		
		// if email exists, just create branch and user_branches
		err = s.unitOfWork.WithinTransaction(ctx, func(repos domain.RegistrationRepositories) error {
			if err := repos.BranchRepo.CreateBranch(ctx, branch); err != nil {
				return err
			}
			if _, err := repos.UserBranchesRepo.CreateUserBranches(existingUser.ID, branch.ID); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to register branch for existing user: %w", err)
		}
		
	} else {
		// Use the unit of work to create user and branch in a transaction
	
		err = s.unitOfWork.WithinTransaction(ctx, func(repos domain.RegistrationRepositories) error {
			if err := repos.UserRepo.Create(ctx, user); err != nil {
				return err
			}
			if err := repos.BranchRepo.CreateBranch(ctx, branch); err != nil {
				return err
			}
			if _, err := repos.UserBranchesRepo.CreateUserBranches(user.ID, branch.ID); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("failed to register user and branch: %w", err)
		}
	}


	return &domain.RegistrationResult{
		User:   user,
		Branch: branch,
	}, nil
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
