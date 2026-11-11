package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/repository"
)

type inMemoryTicketRepo struct {
	tickets  map[int64]*domain.Ticket
	comments map[int64][]domain.Comment
}

func newInMemoryTicketRepo() *inMemoryTicketRepo {
	return &inMemoryTicketRepo{
		tickets:  make(map[int64]*domain.Ticket),
		comments: make(map[int64][]domain.Comment),
	}
}

func (m *inMemoryTicketRepo) Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	t.ID = int64(len(m.tickets) + 1)
	t.CreatedAt = time.Now()
	t.UpdatedAt = time.Now()
	m.tickets[t.ID] = t
	return t, nil
}

func (m *inMemoryTicketRepo) GetByID(ctx context.Context, id int64) (*domain.Ticket, error) {
	t, ok := m.tickets[id]
	if !ok {
		return nil, domain.ErrTicketNotFound
	}
	cp := *t
	cp.Comments = append([]domain.Comment(nil), m.comments[id]...)
	return &cp, nil
}

func (m *inMemoryTicketRepo) List(ctx context.Context, filter repository.TicketFilter) (*repository.TicketListResult, error) {
	var list []*domain.Ticket
	for _, t := range m.tickets {
		if filter.RequesterID != nil && t.RequesterID != *filter.RequesterID {
			continue
		}
		list = append(list, t)
	}
	return &repository.TicketListResult{
		Tickets:    list,
		TotalCount: int64(len(list)),
		Page:       1,
		Limit:      10,
		TotalPages: 1,
	}, nil
}

func (m *inMemoryTicketRepo) Update(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error) {
	m.tickets[t.ID] = t
	return m.GetByID(ctx, t.ID)
}

func (m *inMemoryTicketRepo) CreateComment(ctx context.Context, c *domain.Comment) (*domain.Comment, error) {
	c.ID = int64(len(m.comments[c.TicketID]) + 1)
	c.CreatedAt = time.Now()
	m.comments[c.TicketID] = append(m.comments[c.TicketID], *c)
	return c, nil
}

type inMemoryUserRepo struct {
	users map[int64]*domain.User
}

func (u *inMemoryUserRepo) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	usr, ok := u.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return usr, nil
}

func TestTicketAccess_IDORProtection(t *testing.T) {
	repo := newInMemoryTicketRepo()
	userRepo := &inMemoryUserRepo{users: make(map[int64]*domain.User)}
	svc := NewTicketService(repo, userRepo, 2)
	ctx := context.Background()

	requesterA := int64(101) // Siti Rahmawati
	requesterB := int64(102) // Dewi Lestari
	agent := int64(201)      // Nadia Putri

	ticket, err := repo.Create(ctx, &domain.Ticket{
		Title:       "Akses VPN kantor terputus setelah update Windows",
		Description: "Koneksi OpenVPN gagal TLS handshake",
		Category:    "Network",
		Priority:    domain.PriorityHigh,
		Status:      domain.StatusOpen,
		RequesterID: requesterA,
	})
	if err != nil {
		t.Fatalf("setup ticket gagal: %v", err)
	}

	// 1. Dewi (requester B) mencoba melihat tiket milik Siti (requester A) -> Harus 403 Forbidden
	_, err = svc.GetTicketByID(ctx, requesterB, domain.RoleRequester, ticket.ID)
	if !errors.Is(err, domain.ErrUnauthorizedTicket) {
		t.Fatalf("expected ErrUnauthorizedTicket saat requester B akses tiket A, got %v", err)
	}

	// 2. Dewi (requester B) mencoba menambah komentar pada tiket milik Siti -> Harus 403 Forbidden
	_, err = svc.AddComment(ctx, requesterB, domain.RoleRequester, ticket.ID, CreateCommentInput{
		Body: "Komentar ilegal dari pengguna lain",
	})
	if !errors.Is(err, domain.ErrUnauthorizedTicket) {
		t.Fatalf("expected ErrUnauthorizedTicket saat requester B komentar di tiket A, got %v", err)
	}

	// 3. Siti (requester A / owner) dapat melihat tiket miliknya sendiri
	detailA, err := svc.GetTicketByID(ctx, requesterA, domain.RoleRequester, ticket.ID)
	if err != nil {
		t.Fatalf("owner harus dapat melihat tiket sendiri, got err: %v", err)
	}
	if detailA.ID != ticket.ID {
		t.Errorf("expected ticket ID %d, got %d", ticket.ID, detailA.ID)
	}

	// 4. Nadia (agent) dapat melihat tiket Siti
	detailAgent, err := svc.GetTicketByID(ctx, agent, domain.RoleAgent, ticket.ID)
	if err != nil {
		t.Fatalf("agent harus dapat melihat tiket, got err: %v", err)
	}
	if detailAgent.ID != ticket.ID {
		t.Errorf("expected ticket ID %d, got %d", ticket.ID, detailAgent.ID)
	}
}

func TestTicketComments_InternalNoteFiltering(t *testing.T) {
	repo := newInMemoryTicketRepo()
	userRepo := &inMemoryUserRepo{users: make(map[int64]*domain.User)}
	svc := NewTicketService(repo, userRepo, 2)
	ctx := context.Background()

	requesterID := int64(101) // Siti Rahmawati
	agentID := int64(201)     // Nadia Putri

	ticket, err := repo.Create(ctx, &domain.Ticket{
		Title:       "Layar monitor eksternal ruang meeting Lt. 3 berkedip",
		Description: "Monitor blank setiap 5 detik",
		Category:    "Hardware",
		Priority:    domain.PriorityUrgent,
		Status:      domain.StatusOpen,
		RequesterID: requesterID,
	})
	if err != nil {
		t.Fatalf("setup ticket gagal: %v", err)
	}

	// Tambahkan 1 komentar publik
	_, err = svc.AddComment(ctx, requesterID, domain.RoleRequester, ticket.ID, CreateCommentInput{
		Body:       "Sudah coba ganti colokan listrik tetap kedip",
		IsInternal: false,
	})
	if err != nil {
		t.Fatalf("tambah komentar publik gagal: %v", err)
	}

	// Tambahkan 1 internal note oleh agent
	_, err = svc.AddComment(ctx, agentID, domain.RoleAgent, ticket.ID, CreateCommentInput{
		Body:       "Root CA cert di config VPN klien expired, cek inventaris kabel HDMI Lt. 3",
		IsInternal: true,
	})
	if err != nil {
		t.Fatalf("tambah internal note gagal: %v", err)
	}

	// Requester mencoba membuat internal note -> Harus ditolak (403)
	_, err = svc.AddComment(ctx, requesterID, domain.RoleRequester, ticket.ID, CreateCommentInput{
		Body:       "Percobaan requester membuat catatan rahasia",
		IsInternal: true,
	})
	if !errors.Is(err, domain.ErrUnauthorizedTicket) {
		t.Fatalf("expected ErrUnauthorizedTicket saat requester mencoba membuat internal note, got %v", err)
	}

	// Requester mengambil detail tiket: Komentar internal HARUS disaring
	requesterView, err := svc.GetTicketByID(ctx, requesterID, domain.RoleRequester, ticket.ID)
	if err != nil {
		t.Fatalf("ambil detail tiket oleh requester gagal: %v", err)
	}

	for _, c := range requesterView.Comments {
		if c.IsInternal {
			t.Fatalf("keamanan bocor: komentar internal terlihat oleh requester: %+v", c)
		}
	}
	if len(requesterView.Comments) != 1 {
		t.Fatalf("expected 1 public comment untuk requester, got %d", len(requesterView.Comments))
	}

	// Agent mengambil detail tiket: Semua komentar (publik dan internal) HARUS terlihat
	agentView, err := svc.GetTicketByID(ctx, agentID, domain.RoleAgent, ticket.ID)
	if err != nil {
		t.Fatalf("ambil detail tiket oleh agent gagal: %v", err)
	}
	if len(agentView.Comments) != 2 {
		t.Fatalf("expected 2 comments untuk agent (publik + internal), got %d", len(agentView.Comments))
	}
}

func TestTicket_SLABreachCalculation(t *testing.T) {
	now := time.Now().UTC()

	tests := []struct {
		name           string
		priority       domain.Priority
		status         domain.Status
		createdAt      time.Time
		urgentHours    int
		expectBreach   bool
		expectDeadline bool
	}{
		{
			name:           "urgent open ticket older than SLA limit is breached",
			priority:       domain.PriorityUrgent,
			status:         domain.StatusOpen,
			createdAt:      now.Add(-3 * time.Hour),
			urgentHours:    2,
			expectBreach:   true,
			expectDeadline: true,
		},
		{
			name:           "urgent open ticket newer than SLA limit is not breached",
			priority:       domain.PriorityUrgent,
			status:         domain.StatusOpen,
			createdAt:      now.Add(-1 * time.Hour),
			urgentHours:    2,
			expectBreach:   false,
			expectDeadline: true,
		},
		{
			name:           "urgent in_progress ticket is not breached",
			priority:       domain.PriorityUrgent,
			status:         domain.StatusInProgress,
			createdAt:      now.Add(-5 * time.Hour),
			urgentHours:    2,
			expectBreach:   false,
			expectDeadline: true,
		},
		{
			name:           "urgent resolved ticket is not breached",
			priority:       domain.PriorityUrgent,
			status:         domain.StatusResolved,
			createdAt:      now.Add(-5 * time.Hour),
			urgentHours:    2,
			expectBreach:   false,
			expectDeadline: true,
		},
		{
			name:           "high priority open ticket is not subject to urgent SLA",
			priority:       domain.PriorityHigh,
			status:         domain.StatusOpen,
			createdAt:      now.Add(-10 * time.Hour),
			urgentHours:    2,
			expectBreach:   false,
			expectDeadline: false,
		},
		{
			name:           "medium priority open ticket is not subject to urgent SLA",
			priority:       domain.PriorityMedium,
			status:         domain.StatusOpen,
			createdAt:      now.Add(-24 * time.Hour),
			urgentHours:    2,
			expectBreach:   false,
			expectDeadline: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ticket := &domain.Ticket{
				Priority:  tc.priority,
				Status:    tc.status,
				CreatedAt: tc.createdAt,
			}
			ticket.CalculateSLA(tc.urgentHours, now)

			if ticket.IsSLABreached != tc.expectBreach {
				t.Errorf("expected IsSLABreached = %v, got %v", tc.expectBreach, ticket.IsSLABreached)
			}
			if tc.expectDeadline && ticket.SLADeadline == nil {
				t.Errorf("expected SLADeadline to be set")
			}
			if !tc.expectDeadline && ticket.SLADeadline != nil {
				t.Errorf("expected SLADeadline to be nil, got %v", ticket.SLADeadline)
			}
		})
	}
}
