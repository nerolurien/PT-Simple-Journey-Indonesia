package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/repository"
	"simplejourney/helpdesk/backend/internal/service"
)

type TicketHandler struct {
	ticketService *service.TicketService
}

func NewTicketHandler(ticketService *service.TicketService) *TicketHandler {
	return &TicketHandler{ticketService: ticketService}
}

func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	var input service.CreateTicketInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	ticket, err := h.ticketService.CreateTicket(r.Context(), userID, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) || errors.Is(err, domain.ErrInvalidPriority) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal membuat tiket")
		return
	}

	writeJSON(w, http.StatusCreated, ticket)
}

func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	var status *domain.Status
	if s := strings.TrimSpace(q.Get("status")); s != "" {
		st := domain.Status(s)
		if st.IsValid() {
			status = &st
		}
	}

	var priority *domain.Priority
	if p := strings.TrimSpace(q.Get("priority")); p != "" {
		pr := domain.Priority(p)
		if pr.IsValid() {
			priority = &pr
		}
	}

	var assigneeID *int64
	if a := strings.TrimSpace(q.Get("assignee_id")); a != "" {
		if id, err := strconv.ParseInt(a, 10, 64); err == nil {
			assigneeID = &id
		}
	}

	filter := repository.TicketFilter{
		Status:     status,
		Priority:   priority,
		AssigneeID: assigneeID,
		Search:     q.Get("q"),
		SortBy:     q.Get("sort_by"),
		Order:      q.Get("order"),
		Page:       page,
		Limit:      limit,
	}

	result, err := h.ticketService.ListTickets(r.Context(), userID, role, filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal mengambil daftar tiket")
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	idStr := chi.URLParam(r, "id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tiket tidak valid")
		return
	}

	ticket, err := h.ticketService.GetTicketByID(r.Context(), userID, role, ticketID)
	if err != nil {
		if errors.Is(err, domain.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "tiket tidak ditemukan")
			return
		}
		if errors.Is(err, domain.ErrUnauthorizedTicket) {
			writeError(w, http.StatusForbidden, "akses tiket tidak diizinkan")
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal mengambil detail tiket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	idStr := chi.URLParam(r, "id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tiket tidak valid")
		return
	}

	var input service.UpdateTicketInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	ticket, err := h.ticketService.UpdateTicket(r.Context(), userID, role, ticketID, input)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidTransition) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, domain.ErrForbiddenTransition) || errors.Is(err, domain.ErrUnauthorizedTicket) {
			writeError(w, http.StatusForbidden, err.Error())
			return
		}
		if errors.Is(err, domain.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "tiket tidak ditemukan")
			return
		}
		if errors.Is(err, domain.ErrInvalidPriority) || errors.Is(err, service.ErrInvalidAssignee) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal memperbarui tiket")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

func (h *TicketHandler) AddComment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	idStr := chi.URLParam(r, "id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tiket tidak valid")
		return
	}

	var input service.CreateCommentInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	comment, err := h.ticketService.AddComment(r.Context(), userID, role, ticketID, input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorizedTicket) {
			writeError(w, http.StatusForbidden, "akses komentar tiket tidak diizinkan")
			return
		}
		if errors.Is(err, domain.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "tiket tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal menambahkan komentar")
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *TicketHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	idStr := chi.URLParam(r, "id")
	ticketID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tiket tidak valid")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20+1024)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeError(w, http.StatusUnprocessableEntity, "ukuran file melebihi batas maksimal 5 MB")
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file lampiran wajib diunggah dalam field 'file'")
		return
	}
	defer file.Close()

	att, err := h.ticketService.UploadAttachment(r.Context(), userID, role, ticketID, fileHeader.Filename, fileHeader.Size, file)
	if err != nil {
		if errors.Is(err, domain.ErrFileTooLarge) || errors.Is(err, domain.ErrInvalidFileType) {
			writeError(w, http.StatusUnprocessableEntity, err.Error())
			return
		}
		if errors.Is(err, domain.ErrUnauthorizedTicket) {
			writeError(w, http.StatusForbidden, "akses tiket tidak diizinkan")
			return
		}
		if errors.Is(err, domain.ErrTicketNotFound) {
			writeError(w, http.StatusNotFound, "tiket tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal mengunggah lampiran")
		return
	}

	writeJSON(w, http.StatusCreated, att)
}

func (h *TicketHandler) DownloadAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}
	role, ok := middleware.GetUserRole(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	ticketID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID tiket tidak valid")
		return
	}

	attachmentID, err := strconv.ParseInt(chi.URLParam(r, "att_id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID lampiran tidak valid")
		return
	}

	att, filePath, err := h.ticketService.DownloadAttachment(r.Context(), userID, role, ticketID, attachmentID)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorizedTicket) {
			writeError(w, http.StatusForbidden, "akses lampiran tidak diizinkan")
			return
		}
		if errors.Is(err, domain.ErrTicketNotFound) || errors.Is(err, domain.ErrAttachmentNotFound) {
			writeError(w, http.StatusNotFound, "lampiran tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal mengunduh lampiran")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", att.OriginalName))
	w.Header().Set("Content-Type", att.MimeType)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.ServeFile(w, r, filePath)
}

func (h *TicketHandler) GetAdminStats(w http.ResponseWriter, r *http.Request) {
	role, ok := middleware.GetUserRole(r.Context())
	if !ok || role != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "hanya admin yang dapat mengakses statistik")
		return
	}

	stats, err := h.ticketService.GetAdminStats(r.Context(), role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal memuat data statistik admin")
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
