package httptest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pos-backend/internal/domain"
	handler "pos-backend/internal/handler/http"
)

type MockTicketService struct {
	CreateTicketFunc func(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error)
}

func (m *MockTicketService) CreateTicket(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error) {
	if m.CreateTicketFunc != nil {
		return m.CreateTicketFunc(ctx, branchId, serviceId, userId, items, payments)
	}
	return nil, errors.New("not implemented")
}

func TestTicketHandler_CreateTicket_Scenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		requestBody    interface{}
		mockCreate     func(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error)
		expectedStatus int
		assertBody     func(t *testing.T, bodyBytes []byte)
	}{
		{
			name: "Success Create Ticket",
			requestBody: map[string]interface{}{
				"branch_id":  1,
				"service_id": 2,
				"user_id":    1,
				"items": []map[string]interface{}{
					{"menu_id": 1, "quantity": 2},
				},
				"payments": []map[string]interface{}{
					{"payment_type_id": 1, "amount": 50000.0},
				},
			},
			mockCreate: func(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error) {
				return &domain.Ticket{
					ID:           1,
					BranchId:     branchId,
					ServiceId:    serviceId,
					UserId:       userId,
					TicketNumber: "TKT-1-001",
					TotalAmount:  50000.0,
					Status:       "PENDING",
				}, nil
			},
			expectedStatus: http.StatusCreated,
			assertBody: func(t *testing.T, bodyBytes []byte) {
				var res map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &res); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if res["status"] != "success" {
					t.Errorf("expected status 'success', got %v", res["status"])
				}
				if res["data"] == nil {
					t.Errorf("expected ticket data in response")
				}
			},
		},
		{
			name:           "Invalid Request Payload",
			requestBody:    "malformed-json",
			mockCreate:     nil,
			expectedStatus: http.StatusBadRequest,
			assertBody:     nil,
		},
		{
			name: "Service Error - Internal Server Error",
			requestBody: map[string]interface{}{
				"branch_id":  1,
				"service_id": 2,
				"user_id":    1,
				"items": []map[string]interface{}{
					{"menu_id": 1, "quantity": 1},
				},
				"payments": []map[string]interface{}{
					{"payment_type_id": 1, "amount": 25000.0},
				},
			},
			mockCreate: func(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error) {
				return nil, errors.New("database transaction failed")
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody:     nil,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockTicketService{
				CreateTicketFunc: tc.mockCreate,
			}
			h := handler.NewTicketHandler(mockService)

			var bodyReader *bytes.Buffer
			if str, ok := tc.requestBody.(string); ok {
				bodyReader = bytes.NewBufferString(str)
			} else {
				jsonBytes, err := json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				bodyReader = bytes.NewBuffer(jsonBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			h.CreateTicket(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.assertBody != nil {
				tc.assertBody(t, rec.Body.Bytes())
			}
		})
	}
}
