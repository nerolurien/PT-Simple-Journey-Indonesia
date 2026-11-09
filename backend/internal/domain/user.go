package domain

import "time"

type Role string

const (
	RoleRequester Role = "requester"
	RoleAgent     Role = "agent"
	RoleAdmin     Role = "admin"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleRequester, RoleAgent, RoleAdmin:
		return true
	default:
		return false
	}
}

type User struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
