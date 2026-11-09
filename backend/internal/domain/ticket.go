package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidTransition   = errors.New("alur transisi status tiket tidak valid")
	ErrForbiddenTransition = errors.New("role Anda tidak memiliki izin untuk perubahan status tiket ini")
	ErrTicketNotFound      = errors.New("tiket tidak ditemukan")
	ErrUnauthorizedTicket  = errors.New("akses tiket tidak diizinkan")
	ErrInvalidPriority     = errors.New("prioritas tiket tidak valid")
)

type Status string

const (
	StatusOpen       Status = "open"
	StatusInProgress Status = "in_progress"
	StatusResolved   Status = "resolved"
	StatusClosed     Status = "closed"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusOpen, StatusInProgress, StatusResolved, StatusClosed:
		return true
	default:
		return false
	}
}

type Priority string

const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
	PriorityUrgent Priority = "urgent"
)

func (p Priority) IsValid() bool {
	switch p {
	case PriorityLow, PriorityMedium, PriorityHigh, PriorityUrgent:
		return true
	default:
		return false
	}
}

type Ticket struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Priority    Priority  `json:"priority"`
	Status      Status    `json:"status"`
	RequesterID int64     `json:"requester_id"`
	AssigneeID  *int64    `json:"assignee_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Requester *User     `json:"requester,omitempty"`
	Assignee  *User     `json:"assignee,omitempty"`
	Comments  []Comment `json:"comments,omitempty"`
}

func ValidateStatusTransition(from, to Status, role Role, isOwner bool) error {
	if from == to {
		return nil
	}

	switch {
	case from == StatusOpen && to == StatusInProgress:
		if role == RoleAgent || role == RoleAdmin {
			return nil
		}
		return ErrForbiddenTransition

	case from == StatusInProgress && to == StatusResolved:
		if role == RoleAgent || role == RoleAdmin {
			return nil
		}
		return ErrForbiddenTransition

	case from == StatusResolved && to == StatusClosed:
		if role == RoleAgent || role == RoleAdmin {
			return nil
		}
		if role == RoleRequester && isOwner {
			return nil
		}
		return ErrForbiddenTransition

	case from == StatusResolved && to == StatusOpen:
		if role == RoleRequester && isOwner {
			return nil
		}
		return ErrForbiddenTransition

	case from == StatusOpen && to == StatusClosed:
		if role == RoleAdmin {
			return nil
		}
		if role == RoleRequester && isOwner {
			return nil
		}
		return ErrForbiddenTransition

	default:
		return ErrInvalidTransition
	}
}
