package service

import (
	"context"
	"errors"
	"testing"

	"simplejourney/helpdesk/backend/internal/domain"
)

func TestUserService_UpdateRole_SelfPrevention(t *testing.T) {
	svc := NewUserService(nil)
	err := svc.UpdateRole(context.Background(), 1, 1, domain.RoleAgent)
	if !errors.Is(err, ErrCannotChangeSelfRole) {
		t.Fatalf("expected ErrCannotChangeSelfRole, got %v", err)
	}
}

func TestUserService_UpdateRole_InvalidRole(t *testing.T) {
	svc := NewUserService(nil)
	err := svc.UpdateRole(context.Background(), 1, 2, domain.Role("superadmin"))
	if !errors.Is(err, ErrInvalidRole) {
		t.Fatalf("expected ErrInvalidRole, got %v", err)
	}
}
