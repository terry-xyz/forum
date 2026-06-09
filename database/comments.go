package database

import (
	"database/sql"
	"forum/models"
)

// CreateComment inserts a new comment under a post for the logged-in author.
func CreateComment(db *sql.DB, authorID int, postID int, content string) error {

	// The schema records created_at automatically, so the handler only supplies
	// the user, post, and submitted body.
	query := "INSERT INTO comments (author_id, post_id, content) VALUES (?, ?, ?)"

	_, err := db.Exec(query, authorID, postID, content)
	if err != nil {
		return err
	}

	// The caller does not need the new comment ID because it redirects home.
	return nil
}

// GetCommentsByPostID returns all comments for one post in conversation order.
func GetCommentsByPostID(db *sql.DB, postID int) ([]models.Comment, error) {

	// Oldest-first ordering makes the rendered thread read from top to bottom.
	query := `
		SELECT id, author_id, post_id, content, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at ASC
	`
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}

	// Close the rows when scanning is complete, even on early return.
	defer rows.Close()

	var comments []models.Comment

	// Build a slice of model values for the handler to render.
	for rows.Next() {
		var comment models.Comment

		err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	// Surface any error that occurred while advancing through the result set.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

// CommentExists reports whether a comment row exists for a submitted ID.
func CommentExists(db *sql.DB, commentID int) (bool, error) {
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM comments WHERE id = ?)", commentID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

// DeleteCommentByIDAndAuthorID removes a comment only when the author matches.
func DeleteCommentByIDAndAuthorID(db *sql.DB, commentID int, authorID int) error {

	// Include author_id in the WHERE clause so a guessed comment ID cannot
	// delete another user's comment.
	query := `
		DELETE FROM comments
		WHERE id = ? AND author_id = ?
	`
	result, err := db.Exec(query, commentID, authorID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
