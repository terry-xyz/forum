package models

import "time"

type Comment struct {
	ID        int
	AuthorID  int
	PostID    int
	Content   string
	CreatedAt time.Time
}
