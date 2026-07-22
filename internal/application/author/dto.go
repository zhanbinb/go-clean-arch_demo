package author

import "time"

// CreateInput is the payload for creating a new author.
type CreateInput struct {
	Name  string `json:"name" binding:"required,max=100"`
	Email string `json:"email" binding:"required,email"`
}

// UpdateInput is a partial update.
type UpdateInput struct {
	Name  *string `json:"name,omitempty" binding:"omitempty,max=100"`
	Email *string `json:"email,omitempty" binding:"omitempty,email"`
}

// AuthorDTO is the application-layer view.
type AuthorDTO struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
