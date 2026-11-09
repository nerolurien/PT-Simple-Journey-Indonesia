package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/repository"
)

var (
	ErrInvalidAssignee = errors.New("teknisi penugasan harus memiliki role agent atau admin")
)

type TicketRepository interface {
	Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	GetByID(ctx context.Context, id int64) (*domain.Ticket, error)
	List(ctx context.Context, filter repository.TicketFilter) (*repository.TicketListResult, error)
	Update(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	CreateComment(ctx context.Context, c *domain.Comment) (*domain.Comment, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type TicketService struct {
	ticketRepo TicketRepository
	userRepo   UserRepository
}

func NewTicketService(ticketRepo TicketRepository, userRepo UserRepository) *TicketService {
	return &TicketService{
		ticketRepo: ticketRepo,
		userRepo:   userRepo,
	}
}

type CreateTicketInput struct {
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Category    string          `json:"category"`
	Priority    domain.Priority `json:"priority"`
}

type UpdateTicketInput struct {
	Status     *domain.Status   `json:"status"`
	Priority   *domain.Priority `json:"priority"`
	AssigneeID *int64           `json:"assignee_id"`
}

type CreateCommentInput struct {
	Body       string `json:"body"`
	IsInternal bool   `json:"is_internal"`
}

func (s *TicketService) CreateTicket(ctx context.Context, requesterID int64, input CreateTicketInput) (*domain.Ticket, error) {
	title := strings.TrimSpace(input.Title)
	if title == "" {
		return nil, fmt.Errorf("%w: judul tiket wajib diisi", ErrInvalidInput)
	}

	description := strings.TrimSpace(input.Description)
	if description == "" {
		return nil, fmt.Errorf("%w: deskripsi tiket wajib diisi", ErrInvalidInput)
	}

	category := strings.TrimSpace(input.Category)
	if category == "" {
		return nil, fmt.Errorf("%w: kategori tiket wajib diisi", ErrInvalidInput)
	}

	if !input.Priority.IsValid() {
		return nil, domain.ErrInvalidPriority
	}

	ticket := &domain.Ticket{
		Title:       title,
		Description: description,
		Category:    category,
		Priority:    input.Priority,
		Status:      domain.StatusOpen,
		RequesterID: requesterID,
	}

	return s.ticketRepo.Create(ctx, ticket)
}

func (s *TicketService) ListTickets(ctx context.Context, currentUserID int64, role domain.Role, filter repository.TicketFilter) (*repository.TicketListResult, error) {
	if role == domain.RoleRequester {
		filter.RequesterID = &currentUserID
	}

	return s.ticketRepo.List(ctx, filter)
}

func (s *TicketService) GetTicketByID(ctx context.Context, currentUserID int64, role domain.Role, id int64) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if role == domain.RoleRequester {
		if ticket.RequesterID != currentUserID {
			return nil, domain.ErrUnauthorizedTicket
		}

		sanitizedComments := make([]domain.Comment, 0, len(ticket.Comments))
		for _, c := range ticket.Comments {
			if !c.IsInternal {
				sanitizedComments = append(sanitizedComments, c)
			}
		}
		ticket.Comments = sanitizedComments
	}

	return ticket, nil
}

func (s *TicketService) UpdateTicket(ctx context.Context, currentUserID int64, role domain.Role, ticketID int64, input UpdateTicketInput) (*domain.Ticket, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	isOwner := ticket.RequesterID == currentUserID

	if role == domain.RoleRequester {
		if !isOwner {
			return nil, domain.ErrUnauthorizedTicket
		}

		if input.Priority != nil || input.AssigneeID != nil {
			return nil, domain.ErrUnauthorizedTicket
		}

		if input.Status != nil {
			if err := domain.ValidateStatusTransition(ticket.Status, *input.Status, role, isOwner); err != nil {
				return nil, err
			}
			ticket.Status = *input.Status
		}
	} else {
		if input.Status != nil {
			if err := domain.ValidateStatusTransition(ticket.Status, *input.Status, role, isOwner); err != nil {
				return nil, err
			}
			ticket.Status = *input.Status
		}

		if input.Priority != nil {
			if !input.Priority.IsValid() {
				return nil, domain.ErrInvalidPriority
			}
			ticket.Priority = *input.Priority
		}

		if input.AssigneeID != nil {
			if *input.AssigneeID > 0 {
				assignee, err := s.userRepo.GetByID(ctx, *input.AssigneeID)
				if err != nil {
					return nil, fmt.Errorf("assignee: %w", err)
				}
				if assignee.Role != domain.RoleAgent && assignee.Role != domain.RoleAdmin {
					return nil, ErrInvalidAssignee
				}
				ticket.AssigneeID = input.AssigneeID
			} else {
				ticket.AssigneeID = nil
			}
		}
	}

	updated, err := s.ticketRepo.Update(ctx, ticket)
	if err != nil {
		return nil, err
	}

	if role == domain.RoleRequester {
		sanitizedComments := make([]domain.Comment, 0, len(updated.Comments))
		for _, c := range updated.Comments {
			if !c.IsInternal {
				sanitizedComments = append(sanitizedComments, c)
			}
		}
		updated.Comments = sanitizedComments
	}

	return updated, nil
}

func (s *TicketService) AddComment(ctx context.Context, currentUserID int64, role domain.Role, ticketID int64, input CreateCommentInput) (*domain.Comment, error) {
	body := strings.TrimSpace(input.Body)
	if body == "" {
		return nil, fmt.Errorf("%w: isi komentar wajib diisi", ErrInvalidInput)
	}

	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if role == domain.RoleRequester {
		if ticket.RequesterID != currentUserID {
			return nil, domain.ErrUnauthorizedTicket
		}
		if input.IsInternal {
			return nil, domain.ErrUnauthorizedTicket
		}
	}

	comment := &domain.Comment{
		TicketID:   ticketID,
		AuthorID:   currentUserID,
		Body:       body,
		IsInternal: input.IsInternal && (role == domain.RoleAgent || role == domain.RoleAdmin),
	}

	return s.ticketRepo.CreateComment(ctx, comment)
}

