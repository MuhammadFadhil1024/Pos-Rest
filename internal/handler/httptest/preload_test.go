package httptest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos-backend/internal/domain"
	handler "pos-backend/internal/handler/http"
)

type MockServiceTypeService struct {
	GetBranchServicesFunc    func(ctx context.Context, id int64) ([]*domain.Service, error)
	GetBranchServiceMenuFunc func(ctx context.Context, branch_id int64) ([]*domain.ServiceMenu, error)
}

func (m *MockServiceTypeService) GetBranchServices(ctx context.Context, id int64) ([]*domain.Service, error) {
	if m.GetBranchServicesFunc != nil {
		return m.GetBranchServicesFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *MockServiceTypeService) GetBranchServiceMenu(ctx context.Context, branch_id int64) ([]*domain.ServiceMenu, error) {
	if m.GetBranchServiceMenuFunc != nil {
		return m.GetBranchServiceMenuFunc(ctx, branch_id)
	}
	return nil, nil
}

type MockPaymentService struct {
	GetServicePaymentsFunc func(ctx context.Context, id int64) ([]*domain.Payment, error)
}

func (m *MockPaymentService) GetServicePayments(ctx context.Context, id int64) ([]*domain.Payment, error) {
	if m.GetServicePaymentsFunc != nil {
		return m.GetServicePaymentsFunc(ctx, id)
	}
	return []*domain.Payment{}, nil
}

func TestPreloadHandler_GetPreloadData_Scenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		branchIDParam  string
		mockServices   func(ctx context.Context, id int64) ([]*domain.Service, error)
		mockBranch     func(ctx context.Context, branchId int64) (*domain.Branch, error)
		mockPayments   func(ctx context.Context, id int64) ([]*domain.Payment, error)
		expectedStatus int
		assertBody     func(t *testing.T, bodyBytes []byte)
	}{
		{
			name:          "Success Preload Data",
			branchIDParam: "1",
			mockServices: func(ctx context.Context, id int64) ([]*domain.Service, error) {
				return []*domain.Service{{ID: 1, Name: "Cashier", Icon: "cash", Payments: []*domain.Payment{{ID: 1, ServiceID: 1, BranchID: 1, Name: "CASH"}}}}, nil
			},
			mockBranch: func(ctx context.Context, branchId int64) (*domain.Branch, error) {
				return &domain.Branch{ID: 1, Name: "Main Branch"}, nil
			},
			mockPayments: func(ctx context.Context, id int64) ([]*domain.Payment, error) {
				return []*domain.Payment{{ID: 1, ServiceID: 1, BranchID: 1, Name: "CASH"}}, nil
			},
			expectedStatus: http.StatusOK,
			assertBody: func(t *testing.T, bodyBytes []byte) {
				var res map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &res); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if res["status"] != "success" {
					t.Errorf("expected status 'success', got %v", res["status"])
				}
				data, ok := res["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected data object in response")
				}
				if data["branch"] == nil {
					t.Errorf("expected branch data to be present")
				}
				services, ok := data["services"].([]interface{})
				if !ok || len(services) == 0 {
					t.Fatalf("expected services slice in response")
				}
				svc := services[0].(map[string]interface{})
				if svc["payments"] == nil {
					t.Errorf("expected payments inside service to be present")
				}
			},
		},
		{
			name:           "Invalid Branch ID",
			branchIDParam:  "abc",
			mockServices:   nil,
			mockBranch:     nil,
			mockPayments:   nil,
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, bodyBytes []byte) {
				var res map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &res); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if res["error"] != "ID not valid" {
					t.Errorf("expected error 'ID not valid', got %v", res["error"])
				}
			},
		},
		{
			name:          "Internal Server Error - Services Fails",
			branchIDParam: "1",
			mockServices: func(ctx context.Context, id int64) ([]*domain.Service, error) {
				return nil, errors.New("service db error")
			},
			mockBranch: func(ctx context.Context, branchId int64) (*domain.Branch, error) {
				return &domain.Branch{ID: 1, Name: "Main Branch"}, nil
			},
			mockPayments: func(ctx context.Context, id int64) ([]*domain.Payment, error) {
				return []*domain.Payment{}, nil
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody:     nil,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			mockServiceType := &MockServiceTypeService{
				GetBranchServicesFunc: tc.mockServices,
			}
			mockBranch := &MockBranchService{
				GetUserBranchFunc: tc.mockBranch,
			}
			mockPayment := &MockPaymentService{
				GetServicePaymentsFunc: tc.mockPayments,
			}

			h := handler.NewPreloadHandler(mockServiceType, mockBranch, mockPayment)

			mux := http.NewServeMux()
			mux.HandleFunc("GET /api/v1/preload/{id}", h.GetPreloadData)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/preload/"+tc.branchIDParam, nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			fmt.Printf("TEST [%s] -> Code: %d\n", tc.name, rec.Code)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.assertBody != nil {
				tc.assertBody(t, rec.Body.Bytes())
			}
		})
	}
}
