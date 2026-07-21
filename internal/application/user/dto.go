package user

import "time"

// UserDTO is the application-layer view of a user (no password hash).
type UserDTO struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
