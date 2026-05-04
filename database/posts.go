package database

import (
	"database/sql"
	"forum/models"
)

// CreatePost inserts a post owned by an existing user.
func CreatePost(db *sql.DB, authorID int, title string, content string) error {
	query := "INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)"

	_, err := db.Exec(query, authorID, title, content)
	if err != nil {
		return err
	}

	return nil
}

// GetAllPosts returns every post in database order.
func GetAllPosts(db *sql.DB) ([]models.Post, error) {
	query := "SELECT * FROM posts"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
