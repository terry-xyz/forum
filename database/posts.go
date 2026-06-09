package database

import (
	"database/sql"
	"errors"
	"forum/models"
)

var ErrInvalidCategoryID = errors.New("invalid category id")

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

// CreatePostWithCategories inserts a post and its category links atomically.
func CreatePostWithCategories(db *sql.DB, authorID int, title string, content string, categoryIDs []int) (int, error) {
	uniqueCategoryIDs, ok := uniquePositiveInts(categoryIDs)
	if !ok || len(uniqueCategoryIDs) == 0 {
		return 0, ErrInvalidCategoryID
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	for _, categoryID := range uniqueCategoryIDs {
		var exists int
		if err := tx.QueryRow("SELECT COUNT(*) FROM categories WHERE id = ?", categoryID).Scan(&exists); err != nil {
			return 0, err
		}
		if exists == 0 {
			return 0, ErrInvalidCategoryID
		}
	}

	result, err := tx.Exec("INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)", authorID, title, content)
	if err != nil {
		return 0, err
	}
	insertedID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, categoryID := range uniqueCategoryIDs {
		_, err := tx.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", insertedID, categoryID)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(insertedID), nil
}

func uniquePositiveInts(ids []int) ([]int, bool) {
	seen := make(map[int]struct{}, len(ids))
	uniqueIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, false
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		uniqueIDs = append(uniqueIDs, id)
	}

	return uniqueIDs, true
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

// DeletePostByIDAndAuthorID removes a post and its dependent rows only when the author matches.
func DeletePostByIDAndAuthorID(db *sql.DB, postID int, authorID int) error {
	// A transaction keeps the post, comments, category links, and reactions in
	// sync if any dependent delete fails.
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// If something fails before Commit, undo all deletes from this transaction.
	rollback := func() {
		_ = tx.Rollback()
	}

	var exists int

	// Confirm ownership before deleting dependent rows so unauthorized requests
	// cannot remove comments or reactions by guessing a post ID.
	err = tx.QueryRow(
		"SELECT COUNT(*) FROM posts WHERE id = ? AND author_id = ?",
		postID,
		authorID,
	).Scan(&exists)
	if err != nil {
		rollback()
		return err
	}

	if exists == 0 {
		rollback()
		return sql.ErrNoRows
	}

	// Reactions attached directly to the post must go before the post row disappears.
	deletePostReactionsQuery := `DELETE FROM post_reactions WHERE post_id = ?`
	// Comment reactions reference comments, so delete them before deleting comments.
	deleteCommentReactionsQuery := `
		DELETE FROM comment_reactions
		WHERE comment_id IN (
			SELECT id FROM comments WHERE post_id = ?
		)
	`
	// Comments and category links depend on the post and are removed explicitly
	// so this helper works even without relying on cascade behavior.
	deleteCommentsQuery := `DELETE FROM comments WHERE post_id = ?`
	deletePostCategoriesQuery := `DELETE FROM post_categories WHERE post_id = ?`
	// Include author_id again in the final delete to preserve the ownership guard.
	deletePostQuery := `DELETE FROM posts WHERE id = ? AND author_id = ?`

	_, err = tx.Exec(deletePostReactionsQuery, postID)
	if err != nil {
		rollback()
		return err
	}
	_, err = tx.Exec(deleteCommentReactionsQuery, postID)
	if err != nil {
		rollback()
		return err
	}
	_, err = tx.Exec(deleteCommentsQuery, postID)
	if err != nil {
		rollback()
		return err
	}
	_, err = tx.Exec(deletePostCategoriesQuery, postID)
	if err != nil {
		rollback()
		return err
	}
	_, err = tx.Exec(deletePostQuery, postID, authorID)
	if err != nil {
		rollback()
		return err
	}

	// Commit makes the dependent deletes visible together.
	return tx.Commit()
}
