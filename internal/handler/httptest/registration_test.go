package httptest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"pos-backend/internal/domain"
	handler "pos-backend/internal/handler/http"
)

// registrationResponse mirrors the JSON shape returned by RegisterUser.
type registrationResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    struct {
		User   domain.User   `json:"user"`
		Branch domain.Branch `json:"branch"`
	} `json:"data"`
}

func TestRegistrationHandler_RegisterUser_Scenarios(t *testing.T) {
	validBody := map[string]string{
		"name":         "Branch Pusat",
		"company_name": "PT Maju Jaya",
		"email":        "owner@example.com",
		"password":     "secret123",
		"phone_number": "08123456789",
		"address":      "Jl. Merdeka No. 1",
	}

	scenarios := []struct {
		name                string
		requestBody         interface{} // map -> JSON encoded, string -> raw body
		serviceError        error
		expectedStatus      int
		expectedError       string
		expectServiceCalled bool
		assertInputs        func(t *testing.T, branchInputs []domain.BranchInput, userInputs []domain.UserInput)
		assertResponse      func(t *testing.T, body registrationResponse)
	}{
		{
			name:                "Success Register",
			requestBody:         validBody,
			expectedStatus:      http.StatusCreated,
			expectServiceCalled: true,
			assertInputs: func(t *testing.T, branchInputs []domain.BranchInput, userInputs []domain.UserInput) {
				if len(branchInputs) != 1 {
					t.Fatalf("expected 1 branch input, got %d", len(branchInputs))
				}
				if branchInputs[0].Name != validBody["name"] {
					t.Errorf("branch name: expected %q, got %q", validBody["name"], branchInputs[0].Name)
				}
				if branchInputs[0].CompanyName != validBody["company_name"] {
					t.Errorf("branch company_name: expected %q, got %q", validBody["company_name"], branchInputs[0].CompanyName)
				}
				if branchInputs[0].Address != validBody["address"] {
					t.Errorf("branch address: expected %q, got %q", validBody["address"], branchInputs[0].Address)
				}

				if len(userInputs) != 1 {
					t.Fatalf("expected 1 user input, got %d", len(userInputs))
				}
				if userInputs[0].Email != validBody["email"] {
					t.Errorf("user email: expected %q, got %q", validBody["email"], userInputs[0].Email)
				}
				if userInputs[0].Password != validBody["password"] {
					t.Errorf("user password: expected %q, got %q", validBody["password"], userInputs[0].Password)
				}
				if userInputs[0].PhoneNumber != validBody["phone_number"] {
					t.Errorf("user phone_number: expected %q, got %q", validBody["phone_number"], userInputs[0].PhoneNumber)
				}
			},
			assertResponse: func(t *testing.T, body registrationResponse) {
				if body.Status != "success" {
					t.Errorf("expected status success, got %q", body.Status)
				}
				if body.Message != "Success Register" {
					t.Errorf("expected message %q, got %q", "Success Register", body.Message)
				}
				if body.Data.User.Email != validBody["email"] {
					t.Errorf("response user email: expected %q, got %q", validBody["email"], body.Data.User.Email)
				}
				if body.Data.User.Role != "admin" {
					t.Errorf("response user role: expected admin, got %q", body.Data.User.Role)
				}
				if body.Data.User.PhoneNumber != validBody["phone_number"] {
					t.Errorf("response user phone_number: expected %q, got %q", validBody["phone_number"], body.Data.User.PhoneNumber)
				}
				if body.Data.Branch.CompanyName != validBody["company_name"] {
					t.Errorf("response branch company_name: expected %q, got %q", validBody["company_name"], body.Data.Branch.CompanyName)
				}
			},
		},
		{
			name:                "Invalid Request Body (Malformed JSON)",
			requestBody:         "invalid-json-string",
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "Invalid request payload",
			expectServiceCalled: false,
		},
		{
			name:                "Invalid Request Body (Empty Body)",
			requestBody:         "",
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "Invalid request payload",
			expectServiceCalled: false,
		},
		{
			name:                "Internal Server Error - Service Fails",
			requestBody:         validBody,
			serviceError:        errors.New("duplicate email"),
			expectedStatus:      http.StatusInternalServerError,
			expectedError:       "Failed to register user. duplicate email",
			expectServiceCalled: true,
		},
		{
			name:                "Missing All Required Fields",
			requestBody:         map[string]string{},
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "missing required fields: name, company_name, email, password, phone_number, address",
			expectServiceCalled: false,
		},
		{
			name: "Missing Required Field - Company Name",
			requestBody: map[string]string{
				"name":         "Branch Pusat",
				"email":        "owner@example.com",
				"password":     "secret123",
				"phone_number": "08123456789",
				"address":      "Jl. Merdeka No. 1",
			},
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "missing required fields: company_name",
			expectServiceCalled: false,
		},
		{
			name: "Missing Required Field - Phone Number",
			requestBody: map[string]string{
				"name":         "Branch Pusat",
				"company_name": "PT Maju Jaya",
				"email":        "owner@example.com",
				"password":     "secret123",
				"address":      "Jl. Merdeka No. 1",
			},
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "missing required fields: phone_number",
			expectServiceCalled: false,
		},
		{
			name: "Whitespace-only Required Field",
			requestBody: map[string]string{
				"name":         "   ",
				"company_name": "PT Maju Jaya",
				"email":        "owner@example.com",
				"password":     "secret123",
				"phone_number": "08123456789",
				"address":      "Jl. Merdeka No. 1",
			},
			expectedStatus:      http.StatusBadRequest,
			expectedError:       "missing required fields: name",
			expectServiceCalled: false,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			var (
				capturedBranchInputs []domain.BranchInput
				capturedUserInputs   []domain.UserInput
				serviceCalled        bool
			)

			mockUserService := &MockAuthService{
				RegisterFunc: func(ctx context.Context, branchInputs []domain.BranchInput, userInputs []domain.UserInput) (*domain.RegistrationResult, error) {
					serviceCalled = true
					capturedBranchInputs = branchInputs
					capturedUserInputs = userInputs

					if tc.serviceError != nil {
						return nil, tc.serviceError
					}

					user := &domain.User{ID: 1, Role: "admin"}
					if len(userInputs) > 0 {
						user.Email = userInputs[0].Email
						user.PhoneNumber = userInputs[0].PhoneNumber
					}

					branch := &domain.Branch{ID: 2}
					if len(branchInputs) > 0 {
						branch.Name = branchInputs[0].Name
						branch.CompanyName = branchInputs[0].CompanyName
						branch.Address = branchInputs[0].Address
					}

					return &domain.RegistrationResult{User: user, Branch: branch}, nil
				},
			}
			// BranchService is required by the constructor but never used by RegisterUser.
			mockBranchService := &MockBranchService{}
			h := handler.NewRegistrationHandler(mockBranchService, mockUserService)

			var bodyReader *bytes.Buffer
			if raw, ok := tc.requestBody.(string); ok {
				bodyReader = bytes.NewBufferString(raw)
			} else {
				jsonBytes, err := json.Marshal(tc.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				bodyReader = bytes.NewBuffer(jsonBytes)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bodyReader)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			// Act
			h.RegisterUser(rec, req)

			// Assert status
			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tc.expectedStatus, rec.Code, rec.Body.String())
			}

			// Assert service invocation
			if serviceCalled != tc.expectServiceCalled {
				t.Errorf("expected service called=%v, got %v", tc.expectServiceCalled, serviceCalled)
			}
			if tc.expectServiceCalled {
				if tc.assertInputs != nil {
					tc.assertInputs(t, capturedBranchInputs, capturedUserInputs)
				}
			}

			// Assert error response
			if tc.expectedError != "" {
				var actualBody map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &actualBody); err != nil {
					t.Fatalf("failed to unmarshal error body: %v (body: %s)", err, rec.Body.String())
				}
				if actualBody["status"] != "error" {
					t.Errorf("expected status %q, got %q", "error", actualBody["status"])
				}
				if actualBody["message"] != tc.expectedError {
					t.Errorf("expected message %q, got %q", tc.expectedError, actualBody["message"])
				}
			}

			// Assert success response
			if tc.assertResponse != nil {
				var actualBody registrationResponse
				if err := json.Unmarshal(rec.Body.Bytes(), &actualBody); err != nil {
					t.Fatalf("failed to unmarshal success body: %v (body: %s)", err, rec.Body.String())
				}
				tc.assertResponse(t, actualBody)
			}

			// Password must never be exposed in the response.
			if strings.Contains(rec.Body.String(), `"password"`) {
				t.Errorf("response body must not contain password field: %s", rec.Body.String())
			}
		})
	}
}
