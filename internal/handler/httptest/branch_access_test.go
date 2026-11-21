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
	"pos-backend/internal/handler/http/middleware"
	"pos-backend/pkg/jwtutil"
)

type mockBranchAccessRepository struct {
	hasAccessFunc func(ctx context.Context, userID, branchID int64) (bool, error)

	calls            int
	capturedUserID   int64
	capturedBranchID int64
}

func (m *mockBranchAccessRepository) UserHasBranchAccess(ctx context.Context, userID, branchID int64) (bool, error) {
	m.calls++
	m.capturedUserID = userID
	m.capturedBranchID = branchID
	if m.hasAccessFunc != nil {
		return m.hasAccessFunc(ctx, userID, branchID)
	}
	return true, nil
}

var _ domain.BranchAccessRepository = (*mockBranchAccessRepository)(nil)

func TestRequireBranchAccess_Scenarios(t *testing.T) {
	scenarios := []struct {
		name              string
		claimsValue       interface{} // nil means no claims in context
		extractor         middleware.BranchIDExtractor
		request           func() *http.Request
		denyAccess        bool
		repoErr           error
		expectedStatus    int
		expectedMessage   string
		expectNext        bool
		expectRepoCall    bool
		expectedUserID    int64
		expectedBranch    int64
		checkRestoredBody bool
	}{
		{
			name:        "Allowed - Branch ID From Query",
			claimsValue: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:   middleware.BranchIDFromQuery("branch_id"),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil)
			},
			expectedStatus: http.StatusOK,
			expectNext:     true,
			expectRepoCall: true,
			expectedUserID: 1,
			expectedBranch: 10,
		},
		{
			name:        "Allowed - Branch ID From Path",
			claimsValue: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:   middleware.BranchIDFromPath("id"),
			request: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/api/v1/preload/7", nil)
				req.SetPathValue("id", "7")
				return req
			},
			expectedStatus: http.StatusOK,
			expectNext:     true,
			expectRepoCall: true,
			expectedUserID: 1,
			expectedBranch: 7,
		},
		{
			name:        "Allowed - Branch ID From JSON Body And Body Restored",
			claimsValue: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:   middleware.BranchIDFromJSONBody("branch_id"),
			request: func() *http.Request {
				body := bytes.NewBufferString(`{"branch_id":9,"service_id":3}`)
				req := httptest.NewRequest(http.MethodPost, "/api/v1/tickets", body)
				req.Header.Set("Content-Type", "application/json")
				return req
			},
			expectedStatus:    http.StatusOK,
			expectNext:        true,
			expectRepoCall:    true,
			expectedUserID:    1,
			expectedBranch:    9,
			checkRestoredBody: true,
		},
		{
			name:            "Unauthorized - No Claims",
			claimsValue:     nil,
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil) },
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Unauthorized: invalid or missing token claims",
		},
		{
			name:            "Unauthorized - Invalid Claims Type",
			claimsValue:     "not-a-claims-object",
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil) },
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Unauthorized: invalid or missing token claims",
		},
		{
			name:            "Unauthorized - Missing User ID",
			claimsValue:     &jwtutil.CustomClaims{UserID: 0, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil) },
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Unauthorized: invalid or missing token claims",
		},
		{
			name:            "Bad Request - Missing Branch ID",
			claimsValue:     &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets", nil) },
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid or missing branch_id",
		},
		{
			name:            "Bad Request - Non Numeric Branch ID",
			claimsValue:     &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=abc", nil) },
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid or missing branch_id",
		},
		{
			name:            "Bad Request - Negative Branch ID",
			claimsValue:     &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=-1", nil) },
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid or missing branch_id",
		},
		{
			name:        "Bad Request - Invalid JSON Body",
			claimsValue: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:   middleware.BranchIDFromJSONBody("branch_id"),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewBufferString("not-json"))
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid or missing branch_id",
		},
		{
			name:        "Bad Request - Branch ID Missing From Body",
			claimsValue: &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:   middleware.BranchIDFromJSONBody("branch_id"),
			request: func() *http.Request {
				return httptest.NewRequest(http.MethodPost, "/api/v1/tickets", bytes.NewBufferString(`{"service_id":1}`))
			},
			expectedStatus:  http.StatusBadRequest,
			expectedMessage: "Invalid or missing branch_id",
		},
		{
			name:            "Internal Server Error - Repository Fails",
			claimsValue:     &jwtutil.CustomClaims{UserID: 1, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil) },
			repoErr:         errors.New("db down"),
			expectedStatus:  http.StatusInternalServerError,
			expectedMessage: "Failed to check branch access",
			expectRepoCall:  true,
			expectedUserID:  1,
			expectedBranch:  10,
		},
		{
			name:            "Forbidden - User Has No Access",
			claimsValue:     &jwtutil.CustomClaims{UserID: 3, Role: "cashier"},
			extractor:       middleware.BranchIDFromQuery("branch_id"),
			request:         func() *http.Request { return httptest.NewRequest(http.MethodGet, "/api/v1/tickets?branch_id=10", nil) },
			denyAccess:      true,
			expectedStatus:  http.StatusForbidden,
			expectedMessage: "Forbidden: no access to this branch",
			expectRepoCall:  true,
			expectedUserID:  3,
			expectedBranch:  10,
		},
	}

	for _, tc := range scenarios {
		t.Run(tc.name, func(t *testing.T) {
			// Arrange
			repo := &mockBranchAccessRepository{}
			if tc.denyAccess {
				repo.hasAccessFunc = func(ctx context.Context, userID, branchID int64) (bool, error) {
					return false, nil
				}
			}
			if tc.repoErr != nil {
				repo.hasAccessFunc = func(ctx context.Context, userID, branchID int64) (bool, error) {
					return false, tc.repoErr
				}
			}

			var (
				nextCalled  bool
				gotBranchID int64
			)

			next := func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if id, ok := r.Context().Value(middleware.BranchContextKey).(int64); ok {
					gotBranchID = id
				}
				if tc.checkRestoredBody {
					var payload map[string]interface{}
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Errorf("downstream handler could not decode restored body: %v", err)
					} else if id, ok := payload["branch_id"].(float64); !ok || int64(id) != tc.expectedBranch {
						t.Errorf("downstream handler expected branch_id %d in body, got %v", tc.expectedBranch, payload["branch_id"])
					}
				}
				w.WriteHeader(http.StatusOK)
			}

			h := middleware.RequireBranchAccess(repo, tc.extractor)(next)

			req := tc.request()
			if tc.claimsValue != nil {
				req = req.WithContext(context.WithValue(req.Context(), middleware.UserContextKey, tc.claimsValue))
			}
			rec := httptest.NewRecorder()

			// Act
			h(rec, req)

			// Assert
			if rec.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tc.expectedStatus, rec.Code, rec.Body.String())
			}
			if nextCalled != tc.expectNext {
				t.Errorf("expected next called=%v, got %v", tc.expectNext, nextCalled)
			}
			if tc.expectNext && gotBranchID != tc.expectedBranch {
				t.Errorf("expected branch id %d in context, got %d", tc.expectedBranch, gotBranchID)
			}

			repoCalled := repo.calls > 0
			if repoCalled != tc.expectRepoCall {
				t.Errorf("expected repository called=%v, got %v", tc.expectRepoCall, repoCalled)
			}
			if tc.expectRepoCall {
				if repo.calls != 1 {
					t.Errorf("expected repository called once, got %d", repo.calls)
				}
				if repo.capturedUserID != tc.expectedUserID {
					t.Errorf("expected repository user id %d, got %d", tc.expectedUserID, repo.capturedUserID)
				}
				if repo.capturedBranchID != tc.expectedBranch {
					t.Errorf("expected repository branch id %d, got %d", tc.expectedBranch, repo.capturedBranchID)
				}
			}

			if tc.expectedMessage != "" {
				var body map[string]string
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatalf("failed to unmarshal error body: %v (body: %s)", err, rec.Body.String())
				}
				if body["status"] != "error" {
					t.Errorf("expected status %q, got %q", "error", body["status"])
				}
				if body["message"] != tc.expectedMessage {
					t.Errorf("expected message %q, got %q", tc.expectedMessage, body["message"])
				}
			}
		})
	}
}
