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

// MockAuthService is a mock implementation of domain.AuthService for scenario testing.
type MockAuthService struct {
	LoginFunc    func(ctx context.Context, email, password string) (string, error)
	RegisterFunc func(ctx context.Context, email, password, role string) (*domain.User, error)
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (string, error) {
	if m.LoginFunc != nil {
		return m.LoginFunc(ctx, email, password)
	}
	return "", errors.New("not implemented")
}

func (m *MockAuthService) Register(ctx context.Context, email, password, role string) (*domain.User, error) {
	if m.RegisterFunc != nil {
		return m.RegisterFunc(ctx, email, password, role)
	}
	return nil, errors.New("not implemented")
}

func TestAuthHandler_Login_Scenarios(t *testing.T) {
	scenarios := []struct {
		name           string
		requestBody    interface{}
		mockLogin      func(ctx context.Context, email, password string) (string, error)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Success Login",
			requestBody: map[string]string{
				"email":    "user@example.com",
				"password": "correctpassword",
			},
			mockLogin: func(ctx context.Context, email, password string) (string, error) {
				return "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...", nil
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"status":  "success",
				"message": "Success Login",
				"token":   "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
			},
		},
		{
			name:           "Invalid Request Body (Malformed JSON)",
			requestBody:    "invalid-json-string",
			mockLogin:      nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "Invalid status",
			},
		},
		{
			name: "Unauthorized - Wrong Credentials",
			requestBody: map[string]string{
				"email":    "user@example.com",
				"password": "wrongpassword",
			},
			mockLogin: func(ctx context.Context, email, password string) (string, error) {
				return "", errors.New("wrong email or pasword!")
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "wrong email or pasword!",
			},
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			mockService := &MockAuthService{
				LoginFunc: tc.mockLogin,
			}
			h := handler.NewAuthHandler(mockService)

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

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Act
			h.Login(rec, req)

			// Assert Status
			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			// Assert Response Body
			var actualBody map[string]interface{}
			if err := json.Unmarshal(rec.Body.Bytes(), &actualBody); err != nil {
				t.Fatalf("failed to unmarshal response body: %v, response was: %s", err, rec.Body.String())
			}

			for key, expectedVal := range tc.expectedBody {
				actualVal, exists := actualBody[key]
				if !exists {
					t.Errorf("expected key %q in response body, but not found", key)
					continue
				}
				if actualVal != expectedVal {
					t.Errorf("for key %q: expected %v (%T), got %v (%T)", key, expectedVal, expectedVal, actualVal, actualVal)
				}
			}
		})
	}
}
