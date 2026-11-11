package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/repository"
)

var (
	ErrInvalidAssignee = errors.New("teknisi penugasan harus memiliki role agent atau admin")
)

var allowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".pdf":  true,
	".txt":  true,
	".log":  true,
}

type TicketRepository interface {
	Create(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	GetByID(ctx context.Context, id int64) (*domain.Ticket, error)
	List(ctx context.Context, filter repository.TicketFilter) (*repository.TicketListResult, error)
	Update(ctx context.Context, t *domain.Ticket) (*domain.Ticket, error)
	CreateComment(ctx context.Context, c *domain.Comment) (*domain.Comment, error)
	CreateAttachment(ctx context.Context, att *domain.Attachment) (*domain.Attachment, error)
	GetAttachmentByID(ctx context.Context, ticketID, attachmentID int64) (*domain.Attachment, error)
	GetAttachmentsByTicketID(ctx context.Context, ticketID int64) ([]domain.Attachment, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id int64) (*domain.User, error)
}

type TicketService struct {
	ticketRepo     TicketRepository
	userRepo       UserRepository
	slaUrgentHours int
	storageDir     string
}

func NewTicketService(ticketRepo TicketRepository, userRepo UserRepository, slaUrgentHours int, storageDir string) *TicketService {
	if slaUrgentHours <= 0 {
		slaUrgentHours = 2
	}
	if storageDir == "" {
		storageDir = "./storage/attachments"
	}
	return &TicketService{
		ticketRepo:     ticketRepo,
		userRepo:       userRepo,
		slaUrgentHours: slaUrgentHours,
		storageDir:     storageDir,
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

	created, err := s.ticketRepo.Create(ctx, ticket)
	if err != nil {
		return nil, err
	}
	created.CalculateSLA(s.slaUrgentHours, time.Now().UTC())
	return created, nil
}

func (s *TicketService) ListTickets(ctx context.Context, currentUserID int64, role domain.Role, filter repository.TicketFilter) (*repository.TicketListResult, error) {
	if role == domain.RoleRequester {
		filter.RequesterID = &currentUserID
	}

	res, err := s.ticketRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	for _, t := range res.Tickets {
		t.CalculateSLA(s.slaUrgentHours, now)
	}

	return res, nil
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

	ticket.CalculateSLA(s.slaUrgentHours, time.Now().UTC())

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
			if ticket.Status == domain.StatusOpen && *input.Status == domain.StatusInProgress && ticket.FirstRespondedAt == nil {
				now := time.Now().UTC()
				ticket.FirstRespondedAt = &now
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

	updated.CalculateSLA(s.slaUrgentHours, time.Now().UTC())

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

func ValidateAttachment(originalName string, sizeBytes int64, headerBytes []byte) (string, error) {
	if sizeBytes <= 0 || sizeBytes > 5*1024*1024 {
		return "", domain.ErrFileTooLarge
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	if !allowedExtensions[ext] {
		return "", domain.ErrInvalidFileType
	}

	detectedMime := http.DetectContentType(headerBytes)
	switch ext {
	case ".png":
		if detectedMime != "image/png" {
			return "", domain.ErrInvalidFileType
		}
	case ".jpg", ".jpeg":
		if detectedMime != "image/jpeg" {
			return "", domain.ErrInvalidFileType
		}
	case ".pdf":
		if detectedMime != "application/pdf" {
			return "", domain.ErrInvalidFileType
		}
	case ".txt", ".log":
		if !strings.HasPrefix(detectedMime, "text/plain") {
			return "", domain.ErrInvalidFileType
		}
	default:
		return "", domain.ErrInvalidFileType
	}

	return detectedMime, nil
}

func generateStoredName(ext string) (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes) + ext, nil
}

func (s *TicketService) UploadAttachment(ctx context.Context, currentUserID int64, role domain.Role, ticketID int64, originalName string, sizeBytes int64, fileReader io.Reader) (*domain.Attachment, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, err
	}

	if role == domain.RoleRequester && ticket.RequesterID != currentUserID {
		return nil, domain.ErrUnauthorizedTicket
	}

	if sizeBytes > 5*1024*1024 {
		return nil, domain.ErrFileTooLarge
	}

	headerBytes := make([]byte, 512)
	n, err := io.ReadFull(fileReader, headerBytes)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("read file header: %w", err)
	}

	mimeType, err := ValidateAttachment(originalName, sizeBytes, headerBytes[:n])
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	storedName, err := generateStoredName(ext)
	if err != nil {
		return nil, fmt.Errorf("generate stored name: %w", err)
	}

	if err := os.MkdirAll(s.storageDir, 0750); err != nil {
		return nil, fmt.Errorf("create storage dir: %w", err)
	}

	fullPath := filepath.Join(s.storageDir, storedName)
	dst, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0640)
	if err != nil {
		return nil, fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Write(headerBytes[:n]); err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("write header: %w", err)
	}

	written, err := io.Copy(dst, fileReader)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, fmt.Errorf("copy file body: %w", err)
	}

	totalWritten := int64(n) + written
	if totalWritten > 5*1024*1024 {
		_ = os.Remove(fullPath)
		return nil, domain.ErrFileTooLarge
	}

	cleanOriginalName := filepath.Base(originalName)

	att := &domain.Attachment{
		TicketID:     ticketID,
		UploaderID:   currentUserID,
		OriginalName: cleanOriginalName,
		StoredName:   storedName,
		MimeType:     mimeType,
		SizeBytes:    totalWritten,
	}

	created, err := s.ticketRepo.CreateAttachment(ctx, att)
	if err != nil {
		_ = os.Remove(fullPath)
		return nil, err
	}

	return created, nil
}

func (s *TicketService) DownloadAttachment(ctx context.Context, currentUserID int64, role domain.Role, ticketID, attachmentID int64) (*domain.Attachment, string, error) {
	ticket, err := s.ticketRepo.GetByID(ctx, ticketID)
	if err != nil {
		return nil, "", err
	}

	if role == domain.RoleRequester && ticket.RequesterID != currentUserID {
		return nil, "", domain.ErrUnauthorizedTicket
	}

	att, err := s.ticketRepo.GetAttachmentByID(ctx, ticketID, attachmentID)
	if err != nil {
		return nil, "", err
	}

	fullPath := filepath.Clean(filepath.Join(s.storageDir, att.StoredName))
	if _, err := os.Stat(fullPath); err != nil {
		return nil, "", domain.ErrAttachmentNotFound
	}

	return att, fullPath, nil
}

