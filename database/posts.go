package database

import (
	"database/sql"
	"forum/models"
)

// CreatePost inserts a post owned by an existing user.
func CreatePost(db *sql.DB, authorID int, title string, content string) (int, error) {
	// The database fills created_at automatically, so only author and submitted
	// form fields are inserted here.
	query := "INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)"

	result, err := db.Exec(query, authorID, title, content)
	if err != nil {
		return 0, err
	}

	// The new ID is needed immediately so the caller can attach categories.
	insertedID, err := result.LastInsertId()

	return int(insertedID), nil
}

// GetAllPosts returns every post in database order.
func GetAllPosts(db *sql.DB) ([]models.Post, error) {
	// SELECT * matches the current table order, including created_at.
	query := "SELECT * FROM posts"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	// Always close row iterators so the database connection can be reused.
	defer rows.Close()

	var posts []models.Post

	// Convert each result row into the lightweight model used by handlers.
	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	// rows.Err catches iteration errors that occur after Query succeeds.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostsByCategoryID returns posts that are linked to one category.
func GetPostsByCategoryID(db *sql.DB, categoryID int) ([]models.Post, error) {

	// Join through post_categories because posts and categories are many-to-many.
	query := `
		SELECT p.id, p.author_id, p.title, p.content, p.created_at
		FROM posts p
		JOIN post_categories pc ON pc.post_id = p.id
		WHERE pc.category_id = ?
		ORDER BY p.created_at DESC

	`
	rows, err := db.Query(query, categoryID)
	if err != nil {
		return nil, err
	}

	// Release the result set after all matching posts have been read.
	defer rows.Close()

	var posts []models.Post

	// Scan each joined post into the same model shape as GetAllPosts.
	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	// Return deferred iteration errors as normal database errors.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetPostsByAuthorID returns posts created by one user in newest-first order.
func GetPostsByAuthorID(db *sql.DB, authorID int) ([]models.Post, error) {

	// Filter by author_id because the session user ID maps directly to posts.author_id.
	query := `
		SELECT id, author_id, title, content, created_at
		FROM posts
		WHERE author_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, authorID)
	if err != nil {
		return nil, err
	}

	// Close the filtered cursor so the connection is available for later queries.
	defer rows.Close()

	var posts []models.Post

	// Scan into the same model shape used by the full feed renderer.
	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	// Return delayed cursor errors that can happen after Query succeeds.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

// GetLikedPostsByUserID returns posts the user liked, ordered newest first.
func GetLikedPostsByUserID(db *sql.DB, userID int) ([]models.Post, error) {

	// Join through post_reactions because liked posts are defined by reaction rows.
	query := `
		SELECT p.id, p.author_id, p.title, p.content, p.created_at
		FROM posts p
		JOIN post_reactions pr ON pr.post_id = p.id
		WHERE pr.user_id = ? AND pr.reaction_type = 'like'
		ORDER BY p.created_at DESC
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	// Close the joined cursor so the connection is available for later queries.
	defer rows.Close()

	var posts []models.Post

	// Scan only post columns because reaction metadata is used only for filtering.
	for rows.Next() {
		var post models.Post

		err := rows.Scan(&post.ID, &post.AuthorID, &post.Title, &post.Content, &post.CreatedAt)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	// Return delayed cursor errors that can happen after Query succeeds.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}
