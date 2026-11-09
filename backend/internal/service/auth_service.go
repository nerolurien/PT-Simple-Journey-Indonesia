package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/repository"
)

var (
	ErrInvalidCredentials = errors.New("Email atau password salah")
	ErrInvalidInput       = errors.New("data permintaan tidak valid")
)

type AuthService struct {
	userRepo  *repository.UserRepository
	jwtSecret string
	tokenTTL  time.Duration
}

func NewAuthService(userRepo *repository.UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		jwtSecret: jwtSecret,
		tokenTTL:  24 * time.Hour,
	}
}

type RegisterInput struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthPayload struct {
	Token string       `json:"token"`
	User  *domain.User `json:"user"`
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*AuthPayload, error) {
	name := strings.TrimSpace(input.Name)
	if name == "" {
		return nil, fmt.Errorf("%w: nama wajib diisi", ErrInvalidInput)
	}

	email := strings.ToLower(strings.TrimSpace(input.Email))
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("%w: format email tidak valid", ErrInvalidInput)
	}

	if len(input.Password) < 8 {
		return nil, fmt.Errorf("%w: password minimal 8 karakter", ErrInvalidInput)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("generate password hash: %w", err)
	}

	user := &domain.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
		Role:         domain.RoleRequester,
	}

	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := middleware.GenerateToken(s.jwtSecret, createdUser.ID, createdUser.Role, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthPayload{
		Token: token,
		User:  createdUser,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (*AuthPayload, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || input.Password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := middleware.GenerateToken(s.jwtSecret, user.ID, user.Role, s.tokenTTL)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthPayload{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*domain.User, error) {
	return s.userRepo.GetByID(ctx, userID)
}
