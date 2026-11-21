package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"pos-backend/internal/domain"
	"pos-backend/pkg/jwtutil"
)

// BranchContextKey is the context key under which the authorized branch ID is stored.
const BranchContextKey contextKey = "branch_id"

// ErrBranchIDNotFound is returned by extractors when the branch ID is absent or invalid.
var ErrBranchIDNotFound = errors.New("branch id not found")

// BranchIDExtractor extracts the branch ID that must be authorized from a request.
type BranchIDExtractor func(r *http.Request) (int64, error)

// RequireBranchAccess ensures the authenticated user (claims from RequireAuth)
// has access to the requested branch. It must be wrapped by RequireAuth.
func RequireBranchAccess(
	repo domain.BranchAccessRepository,
	extract BranchIDExtractor,
) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*jwtutil.CustomClaims)
			if !ok || claims == nil || claims.UserID <= 0 {
				writeJSONError(w, http.StatusUnauthorized, "Unauthorized: invalid or missing token claims")
				return
			}

			branchID, err := extract(r)
			if err != nil || branchID <= 0 {
				writeJSONError(w, http.StatusBadRequest, "Invalid or missing branch_id")
				return
			}

			allowed, err := repo.UserHasBranchAccess(r.Context(), claims.UserID, branchID)
			if err != nil {
				writeJSONError(w, http.StatusInternalServerError, "Failed to check branch access")
				return
			}
			if !allowed {
				writeJSONError(w, http.StatusForbidden, "Forbidden: no access to this branch")
				return
			}

			ctx := context.WithValue(r.Context(), BranchContextKey, branchID)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// BranchIDFromPath reads the branch ID from a path parameter, e.g. /preload/{id}.
func BranchIDFromPath(param string) BranchIDExtractor {
	return func(r *http.Request) (int64, error) {
		return parseBranchID(r.PathValue(param))
	}
}

// BranchIDFromQuery reads the branch ID from a query parameter, e.g. ?branch_id=1.
func BranchIDFromQuery(param string) BranchIDExtractor {
	return func(r *http.Request) (int64, error) {
		return parseBranchID(r.URL.Query().Get(param))
	}
}

// BranchIDFromJSONBody reads the branch ID from a JSON body field.
// The body is restored afterwards so the next handler can decode it again.
func BranchIDFromJSONBody(field string) BranchIDExtractor {
	return func(r *http.Request) (int64, error) {
		if r.Body == nil {
			return 0, ErrBranchIDNotFound
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return 0, err
		}
		// Restore the body for downstream handlers.
		r.Body = io.NopCloser(bytes.NewReader(body))

		var payload map[string]json.RawMessage
		if err := json.Unmarshal(body, &payload); err != nil {
			return 0, err
		}

		raw, ok := payload[field]
		if !ok {
			return 0, ErrBranchIDNotFound
		}

		var branchID int64
		if err := json.Unmarshal(raw, &branchID); err != nil {
			return 0, err
		}

		if branchID <= 0 {
			return 0, ErrBranchIDNotFound
		}

		return branchID, nil
	}
}

func parseBranchID(raw string) (int64, error) {
	if raw == "" {
		return 0, ErrBranchIDNotFound
	}

	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, ErrBranchIDNotFound
	}

	return id, nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "error",
		"message": message,
	})
}
