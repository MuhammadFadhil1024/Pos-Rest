package httptest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos-backend/internal/domain"
	handler "pos-backend/internal/handler/http"
	"pos-backend/internal/handler/http/middleware"
	"pos-backend/pkg/jwtutil"
)

type MockBranchService struct {
	GetUserBranchesFunc func(ctx context.Context, id int64) ([]*domain.Branch, error)
	GetUserBranchFunc   func(ctx context.Context, branchId int64) (*domain.Branch, error)
}

func (m *MockBranchService) GetUserBranches(ctx context.Context, id int64) ([]*domain.Branch, error) {
	if m.GetUserBranchesFunc != nil {
		return m.GetUserBranchesFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockBranchService) GetUserBranch(ctx context.Context, branchId int64) (*domain.Branch, error) {
	if m.GetUserBranchFunc != nil {
		return m.GetUserBranchFunc(ctx, branchId)
	}
	return nil, errors.New("not implemented")
}

func TestBranchHandler_GetMyBranches_Scenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		claims         *jwtutil.CustomClaims
		mockGetUser    func(ctx context.Context, id int64) ([]*domain.Branch, error)
		expectedStatus int
	}{
		{
			name:   "Success Get Branches",
			claims: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			mockGetUser: func(ctx context.Context, id int64) ([]*domain.Branch, error) {
				return []*domain.Branch{
					{ID: 1, Name: "Branch Pusat"},
					{ID: 2, Name: "Branch Cabang 1"},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unauthorized - No Claims",
			claims:         nil,
			mockGetUser:    nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:   "Internal Server Error - Service Fails",
			claims: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			mockGetUser: func(ctx context.Context, id int64) ([]*domain.Branch, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockBranchService{
				GetUserBranchesFunc: tc.mockGetUser,
			}
			h := handler.NewBranchHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/branches", nil)
			if tc.claims != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, tc.claims))
			}
			rec := httptest.NewRecorder()

			h.GeMyBranches(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}
		})
	}
}
