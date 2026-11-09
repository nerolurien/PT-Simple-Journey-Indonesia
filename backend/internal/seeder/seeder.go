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

	tickets := []SeedTicket{
		{
			Title:          "Akses VPN kantor terputus setelah update Windows",
			Description:    "Setelah update Windows 11 otomatis semalam, koneksi OpenVPN ke server staging kantor gagal tersambung dengan pesan error TLS handshake failed. Sudah coba restart laptop tetap tidak bisa.",
			Category:       "Network",
			Priority:       domain.PriorityHigh,
			Status:         domain.StatusInProgress,
			RequesterEmail: "requester@simplejourney.co.id",
			AssigneeEmail:  "agent@simplejourney.co.id",
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
			Comments: []SeedComment{
				{
					AuthorEmail: "admin@simplejourney.co.id",
					Body:        "Akun sudah di-unlock dan temporary password dikirimkan via SMS resmi perusahaan. Silakan login dan ubah password segera.",
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

		var ticketID int64
		err := pool.QueryRow(ctx, `
			SELECT id FROM tickets WHERE title = $1 AND requester_id = $2
		`, t.Title, requesterID).Scan(&ticketID)

		if err != nil {
			if err == pgx.ErrNoRows {
				now := time.Now().UTC()
				err = pool.QueryRow(ctx, `
					INSERT INTO tickets (title, description, category, priority, status, requester_id, assignee_id, created_at, updated_at)
					VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8)
					RETURNING id
				`, t.Title, t.Description, t.Category, t.Priority, t.Status, requesterID, assigneeID, now).Scan(&ticketID)
				if err != nil {
					return fmt.Errorf("insert ticket %s: %w", t.Title, err)
				}
			} else {
				return fmt.Errorf("query ticket %s: %w", t.Title, err)
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
