package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simplejourney/helpdesk/backend/internal/domain"
)

func TestJWTGenerationAndValidation(t *testing.T) {
	secret := "test-secret-key-32-bytes-minimum-length"
	userID := int64(42)
	role := domain.RoleAdmin

	token, err := GenerateToken(secret, userID, role, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := ValidateToken(secret, token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("expected userID %d, got %d", userID, claims.UserID)
	}
	if claims.Role != role {
		t.Errorf("expected role %s, got %s", role, claims.Role)
	}
}

func TestRequireRoles(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name         string
		userRole     *domain.Role
		allowedRoles []domain.Role
		wantStatus   int
	}{
		{
			name:         "allowed role succeeds",
			userRole:     ptr(domain.RoleAdmin),
			allowedRoles: []domain.Role{domain.RoleAdmin, domain.RoleAgent},
			wantStatus:   http.StatusOK,
		},
		{
			name:         "disallowed role forbidden",
			userRole:     ptr(domain.RoleRequester),
			allowedRoles: []domain.Role{domain.RoleAdmin, domain.RoleAgent},
			wantStatus:   http.StatusForbidden,
		},
		{
			name:         "unauthenticated returns unauthorized",
			userRole:     nil,
			allowedRoles: []domain.Role{domain.RoleAdmin},
			wantStatus:   http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.userRole != nil {
				ctx := context.WithValue(req.Context(), roleKey, *tt.userRole)
				req = req.WithContext(ctx)
			}
			rec := httptest.NewRecorder()

			handler := RequireRoles(tt.allowedRoles...)(nextHandler)
			handler.ServeHTTP(rec, req)

			if rec.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, rec.Code)
			}
		})
	}
}

func TestRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(1, 2)
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := RateLimit(limiter)(nextHandler)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d expected status 200, got %d", i, rec.Code)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected rate limit status 429, got %d", rec.Code)
	}
}

func ptr[T any](v T) *T {
	return &v
}
