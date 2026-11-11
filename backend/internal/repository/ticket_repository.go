package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"simplejourney/helpdesk/backend/internal/domain"
)

var validSortColumns = map[string]string{
	"created_at": "t.created_at",
	"updated_at": "t.updated_at",
	"priority":   "t.priority",
}

type TicketFilter struct {
	RequesterID *int64
	AssigneeID  *int64
	Status      *domain.Status
	Priority    *domain.Priority
	Search      string
	SortBy      string
	Order       string
	Page        int
	Limit       int
}

type TicketListResult struct {
	Tickets    []*domain.Ticket `json:"tickets"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	TotalPages int              `json:"total_pages"`
}

type TicketRepository struct {
	pool *pgxpool.Pool
}

func NewTicketRepository(pool *pgxpool.Pool) *TicketRepository {
	return &TicketRepository{pool: pool}
}

func (r *TicketRepository) Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	query := `
		INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id, first_responded_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now().UTC()
	err := r.pool.QueryRow(ctx, query,
		t.Title,
		t.Description,
		t.Category,
		t.Priority,
		t.Status,
		t.RequesterID,
		t.AssigneeID,
		t.FirstRespondedAt,
		now,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("create ticket: %w", err)
	}

	return r.GetByID(ctx, t.ID)
}

func (r *TicketRepository) GetByID(ctx context.Context, id int64) (*domain.Ticket, error) {
	query := `
		SELECT
			t.id, t.title, t.description, t.category, t.priority, t.status,
			t.requester_id, t.assignee_id, t.first_responded_at, t.created_at, t.updated_at,
			req.id, req.name, req.email, req.role,
			asg.id, asg.name, asg.email, asg.role
		FROM tickets t
		JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		WHERE t.id = $1
	`

	var t domain.Ticket
	var reqUser domain.User
	var asgID *int64
	var asgName, asgEmail *string
	var asgRole *domain.Role

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
		&t.RequesterID, &t.AssigneeID, &t.FirstRespondedAt, &t.CreatedAt, &t.UpdatedAt,
		&reqUser.ID, &reqUser.Name, &reqUser.Email, &reqUser.Role,
		&asgID, &asgName, &asgEmail, &asgRole,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTicketNotFound
		}
		return nil, fmt.Errorf("get ticket by id: %w", err)
	}

	t.Requester = &reqUser
	if asgID != nil && asgName != nil && asgEmail != nil && asgRole != nil {
		t.Assignee = &domain.User{
			ID:    *asgID,
			Name:  *asgName,
			Email: *asgEmail,
			Role:  *asgRole,
		}
	}

	comments, err := r.GetCommentsByTicketID(ctx, t.ID)
	if err != nil {
		return nil, fmt.Errorf("get ticket comments: %w", err)
	}
	t.Comments = comments

	return &t, nil
}

func (r *TicketRepository) GetCommentsByTicketID(ctx context.Context, ticketID int64) ([]domain.Comment, error) {
	query := `
		SELECT
			c.id, c.ticket_id, c.author_id, c.body, c.is_internal, c.created_at,
			u.id, u.name, u.email, u.role
		FROM comments c
		JOIN users u ON c.author_id = u.id
		WHERE c.ticket_id = $1
		ORDER BY c.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, ticketID)
	if err != nil {
		return nil, fmt.Errorf("query ticket comments: %w", err)
	}
	defer rows.Close()

	comments := make([]domain.Comment, 0)
	for rows.Next() {
		var c domain.Comment
		var author domain.User
		if err := rows.Scan(
			&c.ID, &c.TicketID, &c.AuthorID, &c.Body, &c.IsInternal, &c.CreatedAt,
			&author.ID, &author.Name, &author.Email, &author.Role,
		); err != nil {
			return nil, fmt.Errorf("scan comment row: %w", err)
		}
		c.Author = &author
		comments = append(comments, c)
	}

	return comments, rows.Err()
}

func (r *TicketRepository) List(ctx context.Context, filter TicketFilter) (*TicketListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 || filter.Limit > 100 {
		filter.Limit = 10
	}

	var conditions []string
	var args []any
	argIndex := 1

	if filter.RequesterID != nil {
		conditions = append(conditions, fmt.Sprintf("t.requester_id = $%d", argIndex))
		args = append(args, *filter.RequesterID)
		argIndex++
	}

	if filter.AssigneeID != nil {
		conditions = append(conditions, fmt.Sprintf("t.assignee_id = $%d", argIndex))
		args = append(args, *filter.AssigneeID)
		argIndex++
	}

	if filter.Status != nil && (*filter.Status).IsValid() {
		conditions = append(conditions, fmt.Sprintf("t.status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	if filter.Priority != nil && (*filter.Priority).IsValid() {
		conditions = append(conditions, fmt.Sprintf("t.priority = $%d", argIndex))
		args = append(args, *filter.Priority)
		argIndex++
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		conditions = append(conditions, fmt.Sprintf("t.title ILIKE $%d", argIndex))
		args = append(args, fmt.Sprintf("%%%s%%", search))
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets t %s", whereClause)
	var totalCount int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, fmt.Errorf("count tickets: %w", err)
	}

	sortCol, ok := validSortColumns[filter.SortBy]
	if !ok {
		sortCol = "t.created_at"
	}

	sortOrder := "DESC"
	if strings.EqualFold(filter.Order, "asc") {
		sortOrder = "ASC"
	}

	offset := (filter.Page - 1) * filter.Limit
	limitArg := argIndex
	offsetArg := argIndex + 1

	selectQuery := fmt.Sprintf(`
		SELECT
			t.id, t.title, t.description, t.category, t.priority, t.status,
			t.requester_id, t.assignee_id, t.first_responded_at, t.created_at, t.updated_at,
			req.id, req.name, req.email, req.role,
			asg.id, asg.name, asg.email, asg.role
		FROM tickets t
		JOIN users req ON t.requester_id = req.id
		LEFT JOIN users asg ON t.assignee_id = asg.id
		%s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortCol, sortOrder, limitArg, offsetArg)

	queryArgs := append(args, filter.Limit, offset)
	rows, err := r.pool.Query(ctx, selectQuery, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list tickets: %w", err)
	}
	defer rows.Close()

	tickets := make([]*domain.Ticket, 0)
	for rows.Next() {
		var t domain.Ticket
		var reqUser domain.User
		var asgID *int64
		var asgName, asgEmail *string
		var asgRole *domain.Role

		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Category, &t.Priority, &t.Status,
			&t.RequesterID, &t.AssigneeID, &t.FirstRespondedAt, &t.CreatedAt, &t.UpdatedAt,
			&reqUser.ID, &reqUser.Name, &reqUser.Email, &reqUser.Role,
			&asgID, &asgName, &asgEmail, &asgRole,
		); err != nil {
			return nil, fmt.Errorf("scan ticket row: %w", err)
		}

		t.Requester = &reqUser
		if asgID != nil && asgName != nil && asgEmail != nil && asgRole != nil {
			t.Assignee = &domain.User{
				ID:    *asgID,
				Name:  *asgName,
				Email: *asgEmail,
				Role:  *asgRole,
			}
		}

		tickets = append(tickets, &t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(filter.Limit)))

	return &TicketListResult{
		Tickets:    tickets,
		TotalCount: totalCount,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *TicketRepository) Update(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	query := `
		UPDATE tickets
		SET status = $1, priority = $2, assignee_id = $3, first_responded_at = $4, updated_at = $5
		WHERE id = $6
	`
	now := time.Now().UTC()
	cmd, err := r.pool.Exec(ctx, query, t.Status, t.Priority, t.AssigneeID, t.FirstRespondedAt, now, t.ID)
	if err != nil {
		return nil, fmt.Errorf("update ticket: %w", err)
	}
	if cmd.RowsAffected() == 0 {
		return nil, domain.ErrTicketNotFound
	}

	return r.GetByID(ctx, t.ID)
}

func (r *TicketRepository) CreateComment(ctx context.Context, c *domain.Comment) (*domain.Comment, error) {
	query := `
		INSERT INTO comments (ticket_id, author_id, body, is_internal, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	now := time.Now().UTC()
	err := r.pool.QueryRow(ctx, query, c.TicketID, c.AuthorID, c.Body, c.IsInternal, now).Scan(&c.ID, &c.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	authorQuery := `
		SELECT id, name, email, role
		FROM users
		WHERE id = $1
	`
	var author domain.User
	if err := r.pool.QueryRow(ctx, authorQuery, c.AuthorID).Scan(&author.ID, &author.Name, &author.Email, &author.Role); err == nil {
		c.Author = &author
	}

	return c, nil
}

