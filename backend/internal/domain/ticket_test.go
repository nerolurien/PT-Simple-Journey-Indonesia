package domain

import (
	"errors"
	"testing"
)

func TestValidateStatusTransition(t *testing.T) {
	tests := []struct {
		name        string
		from        Status
		to          Status
		role        Role
		isOwner     bool
		expectedErr error
	}{
		{
			name:        "same status no-op",
			from:        StatusOpen,
			to:          StatusOpen,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: nil,
		},
		{
			name:        "open to in_progress by agent",
			from:        StatusOpen,
			to:          StatusInProgress,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "open to in_progress by admin",
			from:        StatusOpen,
			to:          StatusInProgress,
			role:        RoleAdmin,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "open to in_progress by requester rejected",
			from:        StatusOpen,
			to:          StatusInProgress,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "in_progress to resolved by agent",
			from:        StatusInProgress,
			to:          StatusResolved,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "in_progress to resolved by admin",
			from:        StatusInProgress,
			to:          StatusResolved,
			role:        RoleAdmin,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "in_progress to resolved by requester rejected",
			from:        StatusInProgress,
			to:          StatusResolved,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "resolved to closed by owner requester",
			from:        StatusResolved,
			to:          StatusClosed,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: nil,
		},
		{
			name:        "resolved to closed by non-owner requester rejected",
			from:        StatusResolved,
			to:          StatusClosed,
			role:        RoleRequester,
			isOwner:     false,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "resolved to closed by agent",
			from:        StatusResolved,
			to:          StatusClosed,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "resolved to closed by admin",
			from:        StatusResolved,
			to:          StatusClosed,
			role:        RoleAdmin,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "resolved to open by owner requester",
			from:        StatusResolved,
			to:          StatusOpen,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: nil,
		},
		{
			name:        "resolved to open by non-owner requester rejected",
			from:        StatusResolved,
			to:          StatusOpen,
			role:        RoleRequester,
			isOwner:     false,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "resolved to open by agent rejected",
			from:        StatusResolved,
			to:          StatusOpen,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "open to closed by owner requester",
			from:        StatusOpen,
			to:          StatusClosed,
			role:        RoleRequester,
			isOwner:     true,
			expectedErr: nil,
		},
		{
			name:        "open to closed by admin",
			from:        StatusOpen,
			to:          StatusClosed,
			role:        RoleAdmin,
			isOwner:     false,
			expectedErr: nil,
		},
		{
			name:        "open to closed by agent rejected",
			from:        StatusOpen,
			to:          StatusClosed,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: ErrForbiddenTransition,
		},
		{
			name:        "closed to anything invalid",
			from:        StatusClosed,
			to:          StatusOpen,
			role:        RoleAdmin,
			isOwner:     false,
			expectedErr: ErrInvalidTransition,
		},
		{
			name:        "open to resolved invalid",
			from:        StatusOpen,
			to:          StatusResolved,
			role:        RoleAgent,
			isOwner:     false,
			expectedErr: ErrInvalidTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusTransition(tt.from, tt.to, tt.role, tt.isOwner)
			if !errors.Is(err, tt.expectedErr) {
				t.Fatalf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
