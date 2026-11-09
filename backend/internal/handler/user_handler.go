package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/repository"
	"simplejourney/helpdesk/backend/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.userService.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "gagal mengambil daftar pengguna")
		return
	}

	writeJSON(w, http.StatusOK, users)
}

type UpdateRoleInput struct {
	Role domain.Role `json:"role"`
}

func (h *UserHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	idStr := chi.URLParam(r, "id")
	targetID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "ID pengguna tidak valid")
		return
	}

	var input UpdateRoleInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	err = h.userService.UpdateRole(r.Context(), adminID, targetID, input.Role)
	if err != nil {
		if errors.Is(err, service.ErrCannotChangeSelfRole) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrInvalidRole) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, repository.ErrUserNotFound) {
			writeError(w, http.StatusNotFound, "pengguna tidak ditemukan")
			return
		}
		writeError(w, http.StatusInternalServerError, "gagal memperbarui role pengguna")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role pengguna berhasil diperbarui"})
}
