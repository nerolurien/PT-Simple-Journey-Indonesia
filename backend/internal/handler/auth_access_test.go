package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/service"
)

type mockUserManagementRepo struct {
	users map[int64]*domain.User
}

func (m *mockUserManagementRepo) List(ctx context.Context) ([]*domain.User, error) {
	var list []*domain.User
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockUserManagementRepo) UpdateRole(ctx context.Context, id int64, role domain.Role) error {
	u, ok := m.users[id]
	if !ok {
		return domain.ErrTicketNotFound
	}
	u.Role = role
	return nil
}

func TestAuthAndRoleUpdateProtection(t *testing.T) {
	jwtSecret := "test-secret-key-32-bytes-minimum-length"

	userRepo := &mockUserManagementRepo{
		users: map[int64]*domain.User{
			1: {ID: 1, Name: "Budi Santoso", Email: "admin@simplejourney.co.id", Role: domain.RoleAdmin},
			2: {ID: 2, Name: "Nadia Putri", Email: "agent@simplejourney.co.id", Role: domain.RoleAgent},
			3: {ID: 3, Name: "Siti Rahmawati", Email: "requester@simplejourney.co.id", Role: domain.RoleRequester},
		},
	}

	userSvc := service.NewUserService(userRepo)
	userHdl := NewUserHandler(userSvc)

	r := chi.NewRouter()
	r.Group(func(admin chi.Router) {
		admin.Use(middleware.Authenticate(jwtSecret))
		admin.Use(middleware.RequireRoles(domain.RoleAdmin))
		admin.Patch("/api/users/{id}/role", userHdl.UpdateRole)
	})

	adminToken, _ := middleware.GenerateToken(jwtSecret, 1, domain.RoleAdmin, time.Hour)
	agentToken, _ := middleware.GenerateToken(jwtSecret, 2, domain.RoleAgent, time.Hour)
	requesterToken, _ := middleware.GenerateToken(jwtSecret, 3, domain.RoleRequester, time.Hour)

	// 1. Unauthenticated request -> 401 Unauthorized
	req := httptest.NewRequest(http.MethodPatch, "/api/users/3/role", bytes.NewBufferString(`{"role":"agent"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized untuk unauthenticated, got %d", rec.Code)
	}

	// 2. Requester tries to update role -> 403 Forbidden
	req = httptest.NewRequest(http.MethodPatch, "/api/users/3/role", bytes.NewBufferString(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+requesterToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden untuk requester, got %d", rec.Code)
	}

	// 3. Agent tries to update role -> 403 Forbidden
	req = httptest.NewRequest(http.MethodPatch, "/api/users/3/role", bytes.NewBufferString(`{"role":"admin"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agentToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden untuk agent, got %d", rec.Code)
	}

	// 4. Admin tries to update own role -> 400 Bad Request
	req = httptest.NewRequest(http.MethodPatch, "/api/users/1/role", bytes.NewBufferString(`{"role":"agent"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request saat admin mengubah role sendiri, got %d", rec.Code)
	}

	// 5. Admin updates other user's role -> 200 OK
	req = httptest.NewRequest(http.MethodPatch, "/api/users/3/role", bytes.NewBufferString(`{"role":"agent"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK saat admin mengubah role user lain, got %d", rec.Code)
	}
	if userRepo.users[3].Role != domain.RoleAgent {
		t.Fatalf("role target gagal terupdate, got %s", userRepo.users[3].Role)
	}
}
