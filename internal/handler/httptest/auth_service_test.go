package httptest

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"pos-backend/internal/domain"
	"pos-backend/internal/service"

	"golang.org/x/crypto/bcrypt"
)

// --- Mocks ---------------------------------------------------------------

type mockUserRepository struct {
	createFunc     func(ctx context.Context, user *domain.User) error
	getByEmailFunc func(ctx context.Context, email string) (*domain.User, error)

	createCalls     int
	createdUser     *domain.User
	getByEmailCalls int
	queriedEmail    string
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	m.createCalls++
	m.createdUser = user
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	if user.ID == 0 {
		user.ID = 1
	}
	return nil
}

func (m *mockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	m.getByEmailCalls++
	m.queriedEmail = email
	if m.getByEmailFunc != nil {
		return m.getByEmailFunc(ctx, email)
	}
	// Default: user does not exist yet (new registration).
	return nil, sql.ErrNoRows
}

type mockBranchRepository struct {
	createFunc        func(ctx context.Context, branch *domain.Branch) error
	getByUserIDFunc   func(ctx context.Context, id int64) ([]*domain.Branch, error)
	getByBranchIDFunc func(ctx context.Context, branchId int64) (*domain.Branch, error)

	createCalls   int
	createdBranch *domain.Branch
}

func (m *mockBranchRepository) GetByUserId(ctx context.Context, id int64) ([]*domain.Branch, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBranchRepository) GetByBranchId(ctx context.Context, branchId int64) (*domain.Branch, error) {
	if m.getByBranchIDFunc != nil {
		return m.getByBranchIDFunc(ctx, branchId)
	}
	return nil, errors.New("not implemented")
}

func (m *mockBranchRepository) CreateBranch(ctx context.Context, branch *domain.Branch) error {
	m.createCalls++
	m.createdBranch = branch
	if m.createFunc != nil {
		return m.createFunc(ctx, branch)
	}
	if branch.ID == 0 {
		branch.ID = 2
	}
	return nil
}

type mockUserBranchesRepository struct {
	createFunc func(userId int64, branchId int64) (*domain.UserBranches, error)

	createCalls      int
	capturedUserID   int64
	capturedBranchID int64
}

func (m *mockUserBranchesRepository) CreateUserBranches(userId int64, branchId int64) (*domain.UserBranches, error) {
	m.createCalls++
	m.capturedUserID = userId
	m.capturedBranchID = branchId
	if m.createFunc != nil {
		return m.createFunc(userId, branchId)
	}
	return &domain.UserBranches{UserId: userId, BranchId: branchId}, nil
}

type mockRegistrationUnitOfWork struct {
	users        domain.UserRepository
	branches     domain.BranchRepository
	userBranches domain.UserBranchesRepository
	withinFunc   func(ctx context.Context, fn func(repos domain.RegistrationRepositories) error) error

	called bool
}

func (m *mockRegistrationUnitOfWork) WithinTransaction(
	ctx context.Context,
	fn func(repos domain.RegistrationRepositories) error,
) error {
	m.called = true
	if m.withinFunc != nil {
		return m.withinFunc(ctx, fn)
	}
	return fn(domain.RegistrationRepositories{
		UserRepo:         m.users,
		BranchRepo:       m.branches,
		UserBranchesRepo: m.userBranches,
	})
}

type registerMocks struct {
	users        *mockUserRepository
	branches     *mockBranchRepository
	userBranches *mockUserBranchesRepository
	uow          *mockRegistrationUnitOfWork
}

func newRegisterMocks() *registerMocks {
	m := &registerMocks{
		users:        &mockUserRepository{},
		branches:     &mockBranchRepository{},
		userBranches: &mockUserBranchesRepository{},
	}
	m.uow = &mockRegistrationUnitOfWork{
		users:        m.users,
		branches:     m.branches,
		userBranches: m.userBranches,
	}
	return m
}

// --- Tests ---------------------------------------------------------------

func TestAuthService_Register_Scenarios(t *testing.T) {
	validBranchInputs := []domain.BranchInput{
		{Name: "Branch Pusat", CompanyName: "PT Maju Jaya", Address: "Jl. Merdeka No. 1"},
	}
	validUserInputs := []domain.UserInput{
		{Email: "owner@example.com", Password: "secret123", PhoneNumber: "08123456789"},
	}

	scenarios := []struct {
		name         string
		branchInputs []domain.BranchInput
		userInputs   []domain.UserInput
		setup        func(m *registerMocks)
		wantErr      string
		check        func(t *testing.T, res *domain.RegistrationResult, m *registerMocks)
	}{
		{
			name:         "Success Register New User",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res == nil {
					t.Fatal("expected registration result, got nil")
				}
				if m.users.getByEmailCalls != 1 {
					t.Errorf("expected GetByEmail to be called once, got %d", m.users.getByEmailCalls)
				}
				if m.users.queriedEmail != validUserInputs[0].Email {
					t.Errorf("expected email lookup %q, got %q", validUserInputs[0].Email, m.users.queriedEmail)
				}
				if !m.uow.called {
					t.Error("expected unit of work to be called")
				}
				if m.users.createCalls != 1 {
					t.Errorf("expected user Create to be called once, got %d", m.users.createCalls)
				}
				if m.branches.createCalls != 1 {
					t.Errorf("expected branch CreateBranch to be called once, got %d", m.branches.createCalls)
				}
				if m.userBranches.createCalls != 1 {
					t.Errorf("expected CreateUserBranches to be called once, got %d", m.userBranches.createCalls)
				}
				if m.userBranches.capturedUserID != 1 || m.userBranches.capturedBranchID != 2 {
					t.Errorf("expected user_branches relation (1,2), got (%d,%d)",
						m.userBranches.capturedUserID, m.userBranches.capturedBranchID)
				}

				if res.User.Email != validUserInputs[0].Email {
					t.Errorf("expected email %q, got %q", validUserInputs[0].Email, res.User.Email)
				}
				if res.User.Role != "admin" {
					t.Errorf("expected default role admin, got %q", res.User.Role)
				}
				if res.User.PhoneNumber != validUserInputs[0].PhoneNumber {
					t.Errorf("expected phone %q, got %q", validUserInputs[0].PhoneNumber, res.User.PhoneNumber)
				}

				if res.User.Password == validUserInputs[0].Password {
					t.Error("password must not be stored as plaintext")
				}
				if err := bcrypt.CompareHashAndPassword([]byte(res.User.Password), []byte(validUserInputs[0].Password)); err != nil {
					t.Errorf("expected stored password to be a valid bcrypt hash of the input: %v", err)
				}

				if res.Branch.Name != validBranchInputs[0].Name {
					t.Errorf("expected branch name %q, got %q", validBranchInputs[0].Name, res.Branch.Name)
				}
				if res.Branch.Slug != "branch-pusat" {
					t.Errorf("expected generated slug %q, got %q", "branch-pusat", res.Branch.Slug)
				}
				if res.Branch.CompanyName != validBranchInputs[0].CompanyName {
					t.Errorf("expected company_name %q, got %q", validBranchInputs[0].CompanyName, res.Branch.CompanyName)
				}
				if res.Branch.Address != validBranchInputs[0].Address {
					t.Errorf("expected address %q, got %q", validBranchInputs[0].Address, res.Branch.Address)
				}
			},
		},
		{
			name:         "No User Inputs",
			branchInputs: validBranchInputs,
			userInputs:   nil,
			wantErr:      "no user inputs provided",
			check: func(t *testing.T, _ *domain.RegistrationResult, m *registerMocks) {
				if m.uow.called {
					t.Error("unit of work must not be called when user inputs are empty")
				}
				if m.users.getByEmailCalls != 0 {
					t.Errorf("GetByEmail must not be called when user inputs are empty, got %d calls", m.users.getByEmailCalls)
				}
			},
		},
		{
			name:         "No Branch Inputs",
			branchInputs: nil,
			userInputs:   validUserInputs,
			wantErr:      "no branch inputs provided",
			check: func(t *testing.T, _ *domain.RegistrationResult, m *registerMocks) {
				if m.uow.called {
					t.Error("unit of work must not be called when branch inputs are empty")
				}
			},
		},
		{
			name:         "Check Existing User Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.users.getByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return nil, errors.New("db down")
				}
			},
			wantErr: "failed to check existing user: db down",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
				if m.uow.called {
					t.Error("unit of work must not be called when the existing-user check fails")
				}
			},
		},
		{
			name:         "User Create Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.users.createFunc = func(ctx context.Context, user *domain.User) error {
					return errors.New("duplicate email")
				}
			},
			wantErr: "failed to register user and branch",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
				if m.branches.createCalls != 0 {
					t.Errorf("branch must not be created when user creation fails, got %d calls", m.branches.createCalls)
				}
				if m.userBranches.createCalls != 0 {
					t.Errorf("relation must not be created when user creation fails, got %d calls", m.userBranches.createCalls)
				}
			},
		},
		{
			name:         "Branch Create Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.branches.createFunc = func(ctx context.Context, branch *domain.Branch) error {
					return errors.New("branch insert error")
				}
			},
			wantErr: "failed to register user and branch",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
				if m.userBranches.createCalls != 0 {
					t.Errorf("relation must not be created when branch creation fails, got %d calls", m.userBranches.createCalls)
				}
			},
		},
		{
			name:         "User Branches Create Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.userBranches.createFunc = func(userId int64, branchId int64) (*domain.UserBranches, error) {
					return nil, errors.New("relation insert error")
				}
			},
			wantErr: "failed to register user and branch",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
			},
		},
		{
			name:         "Transaction Commit Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.uow.withinFunc = func(ctx context.Context, fn func(repos domain.RegistrationRepositories) error) error {
					err := fn(domain.RegistrationRepositories{
						UserRepo:         m.users,
						BranchRepo:       m.branches,
						UserBranchesRepo: m.userBranches,
					})
					if err != nil {
						return err
					}
					return errors.New("commit failed")
				}
			},
			wantErr: "failed to register user and branch: commit failed",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on commit failure")
				}
				if m.users.createCalls != 1 {
					t.Errorf("expected user Create to have been attempted once, got %d", m.users.createCalls)
				}
			},
		},
		{
			name:         "Existing User - Creates Branch Only",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.users.getByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{ID: 7, Email: email, Role: "cashier", PhoneNumber: "0899999999"}, nil
				}
			},
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res == nil {
					t.Fatal("expected registration result, got nil")
				}
				if m.users.createCalls != 0 {
					t.Errorf("user must not be created when email already exists, got %d calls", m.users.createCalls)
				}
				if m.branches.createCalls != 1 {
					t.Errorf("expected branch CreateBranch to be called once, got %d", m.branches.createCalls)
				}
				if m.userBranches.createCalls != 1 {
					t.Errorf("expected CreateUserBranches to be called once, got %d", m.userBranches.createCalls)
				}
				if m.userBranches.capturedUserID != 7 {
					t.Errorf("expected existing user id 7 in relation, got %d", m.userBranches.capturedUserID)
				}
				if m.userBranches.capturedBranchID != 2 {
					t.Errorf("expected new branch id 2 in relation, got %d", m.userBranches.capturedBranchID)
				}
				if res.Branch == nil || res.Branch.Slug != "branch-pusat" {
					t.Errorf("expected returned branch with slug %q, got %+v", "branch-pusat", res.Branch)
				}
			},
		},
		{
			name:         "Existing User - Branch Create Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.users.getByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{ID: 7, Email: email, Role: "cashier"}, nil
				}
				m.branches.createFunc = func(ctx context.Context, branch *domain.Branch) error {
					return errors.New("branch insert error")
				}
			},
			wantErr: "failed to register branch for existing user",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
				if m.userBranches.createCalls != 0 {
					t.Errorf("relation must not be created when branch creation fails, got %d calls", m.userBranches.createCalls)
				}
			},
		},
		{
			name:         "Existing User - Relation Create Fails",
			branchInputs: validBranchInputs,
			userInputs:   validUserInputs,
			setup: func(m *registerMocks) {
				m.users.getByEmailFunc = func(ctx context.Context, email string) (*domain.User, error) {
					return &domain.User{ID: 7, Email: email, Role: "cashier"}, nil
				}
				m.userBranches.createFunc = func(userId int64, branchId int64) (*domain.UserBranches, error) {
					return nil, errors.New("relation insert error")
				}
			},
			wantErr: "failed to register branch for existing user",
			check: func(t *testing.T, res *domain.RegistrationResult, m *registerMocks) {
				if res != nil {
					t.Error("expected nil result on failure")
				}
			},
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			m := newRegisterMocks()
			if tc.setup != nil {
				tc.setup(m)
			}
			svc := service.NewAuthService(m.users, "test-secret", m.uow)

			// Act
			res, err := svc.Register(context.Background(), tc.branchInputs, tc.userInputs)

			// Assert error
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if tc.check != nil {
				tc.check(t, res, m)
			}
		})
	}
}
