package models

import "time"

// Post represents one row from the posts table.
type Post struct {
	ID        int
	AuthorID  int
	Title     string
	Content   string
	CreatedAt time.Time
}
