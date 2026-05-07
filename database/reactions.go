package database

import "database/sql"

// ReactToPost records or replaces a user's reaction on a post.
func ReactToPost(db *sql.DB, userID int, postID int, reactionType string) error {

	// One user can have one reaction per post. The upsert lets a second click
	// change "like" to "dislike" or the other way around.
	query := `
		INSERT INTO post_reactions (user_id, post_id, reaction_type)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, post_id)
		DO UPDATE SET reaction_type = excluded.reaction_type
	`
	_, err := db.Exec(query, userID, postID, reactionType)
	if err != nil {
		return err
	}

	// The handler only needs to know whether the write succeeded.
	return nil
}

// CountPostReactions counts one reaction type for one post.
func CountPostReactions(db *sql.DB, postID int, reactionType string) (int, error) {

	// The reaction type is passed as data, so the same query counts likes and
	// dislikes without building SQL dynamically.
	query := `
		SELECT COUNT(*) FROM post_reactions
		WHERE post_id = ? AND reaction_type = ?	
	`

	var count int

	// COUNT(*) always returns a row, so QueryRow is enough here.
	row := db.QueryRow(query, postID, reactionType)

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// ReactToComment records or replaces a user's reaction on a comment.
func ReactToComment(db *sql.DB, userID int, commentID int, reactionType string) error {

	// The unique key on (user_id, comment_id) keeps each user's comment reaction
	// singular while still allowing them to change their mind.
	query := `
		INSERT INTO comment_reactions (user_id, comment_id, reaction_type)
		VALUES (?, ?, ?)
		ON CONFLICT(user_id, comment_id)
		DO UPDATE SET reaction_type = excluded.reaction_type
	`
	_, err := db.Exec(query, userID, commentID, reactionType)
	if err != nil {
		return err
	}

	// No count is returned because the page reloads and reads fresh totals.
	return nil
}

// CountCommentReactions counts one reaction type for one comment.
func CountCommentReactions(db *sql.DB, commentID int, reactionType string) (int, error) {

	// Reuse one aggregate query for both accepted reaction values.
	query := `
		SELECT COUNT(*) FROM comment_reactions
		WHERE comment_id = ? AND reaction_type = ?
	`

	var count int

	// Scan the aggregate result into a plain integer for display.
	row := db.QueryRow(query, commentID, reactionType)

	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
