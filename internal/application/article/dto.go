package article

import "time"

// CreateInput is what the handler passes to Service.Create.
type CreateInput struct {
	Title    string `json:"title" binding:"required,max=200"`
	Content  string `json:"content" binding:"required"`
	AuthorID int64  `json:"author_id" binding:"required,gt=0"`
}

// UpdateInput is a partial update: nil pointers mean "leave unchanged".
type UpdateInput struct {
	Title   *string `json:"title,omitempty" binding:"omitempty,max=200"`
	Content *string `json:"content,omitempty"`
}

// ArticleDTO is the application-layer view of an article, safe to expose.
type ArticleDTO struct {
	ID         int64     `json:"id"`
	Title      string    `json:"title"`
	Content    string    `json:"content"`
	AuthorID   int64     `json:"author_id"`
	AuthorName string    `json:"author_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ListResult is the paginated list response.
type ListResult struct {
	Items      []*ArticleDTO `json:"items"`
	NextCursor string        `json:"next_cursor,omitempty"`
}
