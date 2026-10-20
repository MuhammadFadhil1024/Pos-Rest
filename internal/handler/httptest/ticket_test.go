package httptest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"pos-backend/internal/domain"
	handler "pos-backend/internal/handler/http"
)

type MockTicketService struct {
	CreateTicketFunc     func(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error)
	GetRecentTicketsFunc func(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error)
}

func (m *MockTicketService) CreateTicket(ctx context.Context, branchId int64, serviceId int64, userId int64, items []domain.TicketItemInput, payments []domain.TicketPaymentInput) (*domain.Ticket, error) {
	if m.CreateTicketFunc != nil {
		return m.CreateTicketFunc(ctx, branchId, serviceId, userId, items, payments)
	}
	return nil, errors.New("not implemented")
}

func (m *MockTicketService) GetRecentTickets(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error) {
	if m.GetRecentTicketsFunc != nil {
		return m.GetRecentTicketsFunc(ctx, filter)
	}
	return nil, 0, errors.New("not implemented")
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

func TestTicketHandler_GetTickets_Scenarios(t *testing.T) {
	now := time.Now()
	sampleTickets := []*domain.TicketListItem{
		{
			ID:           1,
			TicketNumber: "TKT-1-001",
			BranchId:     1,
			UserId:       1,
			CashierName:  "cashier@example.com",
			ServiceId:    2,
			ServiceName:  "DINE IN",
			TotalAmount:  75000.0,
			Status:       "PAID",
			CreatedAt:    now,
		},
	}

	scenarios := []struct {
		name           string
		queryString    string
		mockGetTickets func(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error)
		expectedStatus int
		assertBody     func(t *testing.T, bodyBytes []byte)
	}{
		{
			name:        "Success Get Tickets - Default Pagination",
			queryString: "?branch_id=1&page=1&limit=10",
			mockGetTickets: func(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error) {
				if filter.BranchID != 1 {
					t.Errorf("expected BranchID 1, got %d", filter.BranchID)
				}
				if filter.Limit != 10 {
					t.Errorf("expected Limit 10, got %d", filter.Limit)
				}
				return sampleTickets, 1, nil
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
				data, ok := res["data"].([]interface{})
				if !ok || len(data) != 1 {
					t.Errorf("expected data slice with 1 item, got %v", res["data"])
				}
				meta, ok := res["meta"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected meta object in response")
				}
				if meta["total_items"] != float64(1) {
					t.Errorf("expected total_items 1, got %v", meta["total_items"])
				}
			},
		},
		{
			name:        "Success Get Tickets with Search & Filters",
			queryString: "?branch_id=1&search=001&status=PAID&service_id=2&total_amount=75000",
			mockGetTickets: func(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error) {
				if filter.Search != "001" {
					t.Errorf("expected Search '001', got %s", filter.Search)
				}
				if filter.Status != "PAID" {
					t.Errorf("expected Status 'PAID', got %s", filter.Status)
				}
				if filter.ServiceID != 2 {
					t.Errorf("expected ServiceID 2, got %d", filter.ServiceID)
				}
				if filter.TotalAmount != 75000 {
					t.Errorf("expected TotalAmount 75000, got %f", filter.TotalAmount)
				}
				return sampleTickets, 1, nil
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
			},
		},
		{
			name:           "Bad Request - Missing Branch ID",
			queryString:    "?search=001",
			mockGetTickets: nil,
			expectedStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, bodyBytes []byte) {
				var res map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &res); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if res["error"] == nil {
					t.Errorf("expected error message in response")
				}
			},
		},
		{
			name:        "Internal Server Error - Service Fails",
			queryString: "?branch_id=1",
			mockGetTickets: func(ctx context.Context, filter domain.TicketFilter) ([]*domain.TicketListItem, int64, error) {
				return nil, 0, errors.New("database failure")
			},
			expectedStatus: http.StatusInternalServerError,
			assertBody: func(t *testing.T, bodyBytes []byte) {
				var res map[string]interface{}
				if err := json.Unmarshal(bodyBytes, &res); err != nil {
					t.Fatalf("failed to unmarshal response: %v", err)
				}
				if res["error"] != "Failed to fetch tickets" {
					t.Errorf("expected error 'Failed to fetch tickets', got %v", res["error"])
				}
			},
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockTicketService{
				GetRecentTicketsFunc: tc.mockGetTickets,
			}
			h := handler.NewTicketHandler(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/tickets"+tc.queryString, nil)
			rec := httptest.NewRecorder()

			h.GetTickets(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			if tc.assertBody != nil {
				tc.assertBody(t, rec.Body.Bytes())
			}
		})
	}
}
