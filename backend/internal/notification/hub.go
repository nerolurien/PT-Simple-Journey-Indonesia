package notification

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"simplejourney/helpdesk/backend/internal/domain"
	"simplejourney/helpdesk/backend/internal/middleware"
)

type Event struct {
	Type      string    `json:"type"`
	TicketID  int64     `json:"ticket_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type Client struct {
	UserID int64
	Role   domain.Role
	Send   chan Event
}

type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[*Client]bool),
	}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[c] = true
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.Send)
	}
}

func (h *Hub) BroadcastToRoles(event Event, roles ...domain.Role) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	roleMap := make(map[domain.Role]bool, len(roles))
	for _, r := range roles {
		roleMap[r] = true
	}

	for client := range h.clients {
		if roleMap[client.Role] {
			select {
			case client.Send <- event:
			default:
			}
		}
	}
}

func (h *Hub) SendToUsers(event Event, userIDs ...int64) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userMap := make(map[int64]bool, len(userIDs))
	for _, uid := range userIDs {
		if uid > 0 {
			userMap[uid] = true
		}
	}

	for client := range h.clients {
		if userMap[client.UserID] {
			select {
			case client.Send <- event:
			default:
			}
		}
	}
}

func (h *Hub) HandleStream(jwtSecret string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := ""
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		}
		if tokenStr == "" {
			tokenStr = r.URL.Query().Get("token")
		}

		if tokenStr == "" {
			http.Error(w, "autentikasi diperlukan", http.StatusUnauthorized)
			return
		}

		claims := &middleware.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("algoritma signing tidak valid")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "token tidak valid atau kedaluwarsa", http.StatusUnauthorized)
			return
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming tidak didukung", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")

		client := &Client{
			UserID: claims.UserID,
			Role:   claims.Role,
			Send:   make(chan Event, 32),
		}

		h.Register(client)
		defer h.Unregister(client)

		fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"ok\"}\n\n")
		flusher.Flush()

		ticker := time.NewTicker(25 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				fmt.Fprintf(w, ": ping\n\n")
				flusher.Flush()
			case event, ok := <-client.Send:
				if !ok {
					return
				}
				data, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
				flusher.Flush()
			}
		}
	}
}
