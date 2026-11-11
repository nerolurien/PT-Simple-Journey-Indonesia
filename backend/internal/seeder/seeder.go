package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"simplejourney/helpdesk/backend/internal/domain"
)

type SeedUser struct {
	Name     string
	Email    string
	Password string
	Role     domain.Role
}

type SeedTicket struct {
	Title          string
	Description    string
	Category       string
	Priority       domain.Priority
	Status         domain.Status
	RequesterEmail string
	AssigneeEmail  string
	CreatedAt      *time.Time
	UpdatedAt      *time.Time
	Comments       []SeedComment
}

type SeedComment struct {
	AuthorEmail string
	Body        string
	IsInternal  bool
}

func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	defaultPassword := "Password123!"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	users := []SeedUser{
		{
			Name:     "Budi Santoso",
			Email:    "admin@simplejourney.co.id",
			Password: string(hashedPassword),
			Role:     domain.RoleAdmin,
		},
		{
			Name:     "Nadia Putri",
			Email:    "agent@simplejourney.co.id",
			Password: string(hashedPassword),
			Role:     domain.RoleAgent,
		},
		{
			Name:     "Rizky Pratama",
			Email:    "rizky.pratama@simplejourney.co.id",
			Password: string(hashedPassword),
			Role:     domain.RoleAgent,
		},
		{
			Name:     "Siti Rahmawati",
			Email:    "requester@simplejourney.co.id",
			Password: string(hashedPassword),
			Role:     domain.RoleRequester,
		},
		{
			Name:     "Dewi Lestari",
			Email:    "dewi.lestari@simplejourney.co.id",
			Password: string(hashedPassword),
			Role:     domain.RoleRequester,
		},
	}

	userMap := make(map[string]int64)

	for _, u := range users {
		var id int64
		err := pool.QueryRow(ctx, `
			INSERT INTO users (name, email, password_hash, role)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (email) DO UPDATE
			SET name = EXCLUDED.name, role = EXCLUDED.role
			RETURNING id
		`, u.Name, u.Email, u.Password, u.Role).Scan(&id)
		if err != nil {
			return fmt.Errorf("upsert user %s: %w", u.Email, err)
		}
		userMap[u.Email] = id
	}

	now := time.Now().UTC()
	t1Created := now.Add(-10 * time.Hour)
	t1Updated := now.Add(-2 * time.Hour)

	t2Created := now.Add(-6 * time.Hour)
	t2Updated := now.Add(-6 * time.Hour)

	t3Created := now.Add(-8 * time.Hour)
	t3Updated := now.Add(-2 * time.Hour) // 6 hours resolution

	t4Created := now.Add(-12 * time.Hour)
	t4Updated := now.Add(-9 * time.Hour) // 3 hours resolution (average: 4.5 hours)

	t5Created := now.Add(-5 * time.Hour) // urgent open > 2 hours -> triggers SLA breach count >= 1
	t5Updated := now.Add(-5 * time.Hour)

	tickets := []SeedTicket{
		{
			Title:          "Akses VPN kantor terputus setelah update Windows",
			Description:    "Setelah update Windows 11 otomatis semalam, koneksi OpenVPN ke server staging kantor gagal tersambung dengan pesan error TLS handshake failed. Sudah coba restart laptop tetap tidak bisa.",
			Category:       "Network",
			Priority:       domain.PriorityHigh,
			Status:         domain.StatusInProgress,
			RequesterEmail: "requester@simplejourney.co.id",
			AssigneeEmail:  "agent@simplejourney.co.id",
			CreatedAt:      &t1Created,
			UpdatedAt:      &t1Updated,
			Comments: []SeedComment{
				{
					AuthorEmail: "agent@simplejourney.co.id",
					Body:        `Halo Bu Siti, mohon kirimkan file log OpenVPN terbaru dari folder C:\Program Files\OpenVPN\log agar kami periksa sertifikatnya.`,
					IsInternal:  false,
				},
				{
					AuthorEmail: "requester@simplejourney.co.id",
					Body:        "File log sudah saya kirimkan ke tim IT, terima kasih atas respons cepatnya.",
					IsInternal:  false,
				},
				{
					AuthorEmail: "agent@simplejourney.co.id",
					Body:        "Root CA cert di config VPN klien expired per tanggal 1 bulan ini. Perlu update file ovpn profile untuk user divisi finance.",
					IsInternal:  true,
				},
			},
		},
		{
			Title:          "Request instalasi Docker Desktop di laptop dev",
			Description:    "Membutuhkan approval dan instalasi lisensi Docker Desktop untuk pengerjaan microservices service-billing di laptop workstation baru.",
			Category:       "Software",
			Priority:       domain.PriorityMedium,
			Status:         domain.StatusOpen,
			RequesterEmail: "requester@simplejourney.co.id",
			AssigneeEmail:  "",
			CreatedAt:      &t2Created,
			UpdatedAt:      &t2Updated,
			Comments: []SeedComment{
				{
					AuthorEmail: "requester@simplejourney.co.id",
					Body:        "Sudah mendapat persetujuan dari Tech Lead via email approval terlampir.",
					IsInternal:  false,
				},
			},
		},
		{
			Title:          "Layar monitor eksternal ruang meeting Lt. 3 berkedip",
			Description:    "Monitor Dell 27 inch di meja ruang meeting Lt. 3 sering blank hitam setiap 5 detik saat terhubung ke port HDMI laptop presenter.",
			Category:       "Hardware",
			Priority:       domain.PriorityUrgent,
			Status:         domain.StatusResolved,
			RequesterEmail: "dewi.lestari@simplejourney.co.id",
			AssigneeEmail:  "agent@simplejourney.co.id",
			CreatedAt:      &t3Created,
			UpdatedAt:      &t3Updated,
			Comments: []SeedComment{
				{
					AuthorEmail: "agent@simplejourney.co.id",
					Body:        "Kabel HDMI sudah diganti baru dengan tipe braided 4K. Monitor sudah dites berjalan stabil selama 30 menit.",
					IsInternal:  false,
				},
			},
		},
		{
			Title:          "Reset password akun portal HRIS",
			Description:    "Akun HRIS terkunci karena salah memasukkan password sebanyak 3 kali setelah masa cuti.",
			Category:       "Account & Access",
			Priority:       domain.PriorityLow,
			Status:         domain.StatusClosed,
			RequesterEmail: "dewi.lestari@simplejourney.co.id",
			AssigneeEmail:  "admin@simplejourney.co.id",
			CreatedAt:      &t4Created,
			UpdatedAt:      &t4Updated,
			Comments: []SeedComment{
				{
					AuthorEmail: "admin@simplejourney.co.id",
					Body:        "Akun sudah di-unlock dan temporary password dikirimkan via SMS resmi perusahaan. Silakan login dan ubah password segera.",
					IsInternal:  false,
				},
			},
		},
		{
			Title:          "Database server staging connection pool exhausted",
			Description:    "Layanan backend staging gagal terhubung ke PostgreSQL dengan error max connections reached. Tim QA terblokir untuk verifikasi rilis sprint.",
			Category:       "Database",
			Priority:       domain.PriorityUrgent,
			Status:         domain.StatusOpen,
			RequesterEmail: "dewi.lestari@simplejourney.co.id",
			AssigneeEmail:  "",
			CreatedAt:      &t5Created,
			UpdatedAt:      &t5Updated,
			Comments: []SeedComment{
				{
					AuthorEmail: "dewi.lestari@simplejourney.co.id",
					Body:        "Mohon dibantu restart service atau terminasi idle connection.",
					IsInternal:  false,
				},
			},
		},
	}

	for _, t := range tickets {
		requesterID, ok := userMap[t.RequesterEmail]
		if !ok {
			return fmt.Errorf("requester %s tidak ditemukan di map", t.RequesterEmail)
		}

		var assigneeID *int64
		if t.AssigneeEmail != "" {
			if id, exists := userMap[t.AssigneeEmail]; exists {
				assigneeID = &id
			}
		}

		createdAt := now
		if t.CreatedAt != nil {
			createdAt = *t.CreatedAt
		}
		updatedAt := createdAt
		if t.UpdatedAt != nil {
			updatedAt = *t.UpdatedAt
		}

		var ticketID int64
		err := pool.QueryRow(ctx, `
			SELECT id FROM tickets WHERE title = $1 AND requester_id = $2
		`, t.Title, requesterID).Scan(&ticketID)

		if err != nil {
			if err == pgx.ErrNoRows {
				err = pool.QueryRow(ctx, `
					INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id, first_responded_at, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, NULL, $8, $9)
					RETURNING id
				`, t.Title, t.Description, t.Category, t.Priority, t.Status, requesterID, assigneeID, createdAt, updatedAt).Scan(&ticketID)
				if err != nil {
					return fmt.Errorf("insert ticket %s: %w", t.Title, err)
				}
			} else {
				return fmt.Errorf("query ticket %s: %w", t.Title, err)
			}
		} else {
			_, err = pool.Exec(ctx, `
				UPDATE tickets 
				SET created_at = $1, updated_at = $2, priority = $3, status = $4, assignee_id = $5
				WHERE id = $6
			`, createdAt, updatedAt, t.Priority, t.Status, assigneeID, ticketID)
			if err != nil {
				return fmt.Errorf("update seed ticket %s: %w", t.Title, err)
			}
		}

		for _, c := range t.Comments {
			authorID, ok := userMap[c.AuthorEmail]
			if !ok {
				continue
			}

			var commentID int64
			err := pool.QueryRow(ctx, `
				SELECT id FROM comments WHERE ticket_id = $1 AND author_id = $2 AND body = $3
			`, ticketID, authorID, c.Body).Scan(&commentID)

			if err == pgx.ErrNoRows {
				_, err = pool.Exec(ctx, `
					INSERT INTO comments (ticket_id, author_id, body, is_internal, created_at)
					VALUES ($1, $2, $3, $4, $5)
				`, ticketID, authorID, c.Body, c.IsInternal, time.Now().UTC())
				if err != nil {
					return fmt.Errorf("insert comment for ticket %d: %w", ticketID, err)
				}
			}
		}
	}

	return nil
}
