package models

import "time"

// Comment represents one row from the comments table.
type Comment struct {
	ID        int
	AuthorID  int
	PostID    int
	Content   string
	CreatedAt time.Time
}
