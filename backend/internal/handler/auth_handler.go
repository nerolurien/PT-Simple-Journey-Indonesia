package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"simplejourney/helpdesk/backend/internal/middleware"
	"simplejourney/helpdesk/backend/internal/repository"
	"simplejourney/helpdesk/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input service.RegisterInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	res, err := h.authService.Register(r.Context(), input)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicateEmail) {
			writeError(w, http.StatusConflict, "email sudah terdaftar")
			return
		}
		if errors.Is(err, service.ErrInvalidInput) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
		return
	}

	writeJSON(w, http.StatusCreated, res)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input service.LoginInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "format body request tidak valid")
		return
	}

	res, err := h.authService.Login(r.Context(), input)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "Email atau password salah")
			return
		}
		writeError(w, http.StatusInternalServerError, "terjadi kesalahan internal")
		return
	}

	writeJSON(w, http.StatusOK, res)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "tidak terautentikasi")
		return
	}

	user, err := h.authService.GetCurrentUser(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "pengguna tidak ditemukan")
		return
	}

	writeJSON(w, http.StatusOK, user)
}
