package domain

import (
	"errors"
	"time"
)

var (
	ErrFileTooLarge       = errors.New("ukuran file melebihi batas maksimal 5 MB")
	ErrInvalidFileType    = errors.New("tipe file atau ekstensi tidak diizinkan")
	ErrAttachmentNotFound = errors.New("lampiran tidak ditemukan")
)

type Attachment struct {
	ID           int64     `json:"id"`
	TicketID     int64     `json:"ticket_id"`
	UploaderID   int64     `json:"uploader_id"`
	OriginalName string    `json:"original_name"`
	StoredName   string    `json:"-"`
	MimeType     string    `json:"mime_type"`
	SizeBytes    int64     `json:"size_bytes"`
	CreatedAt    time.Time `json:"created_at"`

	Uploader *User `json:"uploader,omitempty"`
}
