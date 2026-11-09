package service

import (
	"errors"
	"testing"

	"simplejourney/helpdesk/backend/internal/domain"
)

func TestTicketService_RequesterRestrictions(t *testing.T) {
	openStatus := domain.StatusOpen
	closedStatus := domain.StatusClosed
	agentRole := domain.RoleAgent
	requesterRole := domain.RoleRequester

	err := domain.ValidateStatusTransition(openStatus, domain.StatusInProgress, requesterRole, true)
	if !errors.Is(err, domain.ErrForbiddenTransition) {
		t.Fatalf("expected ErrForbiddenTransition, got %v", err)
	}

	err = domain.ValidateStatusTransition(openStatus, domain.StatusInProgress, agentRole, false)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	err = domain.ValidateStatusTransition(closedStatus, openStatus, agentRole, false)
	if !errors.Is(err, domain.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}
