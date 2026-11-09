package service

import (
	"context"
	"errors"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/repository"
)

var (
	ErrCannotChangeSelfRole = errors.New("admin tidak dapat mengubah role akun sendiri")
	ErrInvalidRole          = errors.New("role tidak valid, pilih antara: requester, agent, atau admin")
)

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) ListUsers(ctx context.Context) ([]*domain.User, error) {
	return s.userRepo.List(ctx)
}

func (s *UserService) UpdateRole(ctx context.Context, adminID, targetUserID int64, newRole domain.Role) error {
	if adminID == targetUserID {
		return ErrCannotChangeSelfRole
	}

	if !newRole.IsValid() {
		return ErrInvalidRole
	}

	return s.userRepo.UpdateRole(ctx, targetUserID, newRole)
}
