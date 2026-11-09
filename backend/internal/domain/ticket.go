package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidStatusTransition = errors.New("perubahan status tiket tidak diizinkan")
	ErrTicketNotFound          = errors.New("tiket tidak ditemukan")
	ErrUnauthorizedTicket      = errors.New("akses tiket tidak diizinkan")
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

	Requester *User `json:"requester,omitempty"`
	Assignee  *User `json:"assignee,omitempty"`
}

func CanTransitionStatus(current, next Status, role Role) bool {
	if current == next {
		return true
	}

	if role == RoleRequester {
		return next == StatusClosed
	}

	switch current {
	case StatusOpen:
		return next == StatusInProgress || next == StatusClosed
	case StatusInProgress:
		return next == StatusResolved || next == StatusOpen || next == StatusClosed
	case StatusResolved:
		return next == StatusClosed || next == StatusInProgress
	case StatusClosed:
		return next == StatusOpen
	default:
		return false
	}
}
