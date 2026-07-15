package database

import (
	"database/sql"
	"fmt"
	"forum/models"
	"strings"
)

// ReactionCounts stores like/dislike totals grouped by parent row ID.
type ReactionCounts struct {
	Likes    int
	Dislikes int
}

func placeholders(count int) string {
	if count <= 0 {
		return ""
	}

	return strings.TrimRight(strings.Repeat("?,", count), ",")
}

func intsToArgs(ids []int) []any {
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	return args
}

// GetUsersByIDs loads all requested users in one query.
func GetUsersByIDs(db *sql.DB, ids []int) (map[int]models.User, error) {
	users := make(map[int]models.User, len(ids))
	if len(ids) == 0 {
		return users, nil
	}

	query := fmt.Sprintf("SELECT id, email, username, password FROM users WHERE id IN (%s)", placeholders(len(ids)))
	rows, err := db.Query(query, intsToArgs(ids)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.ID, &user.Email, &user.Username, &user.Password); err != nil {
			return nil, err
		}
		users[user.ID] = user
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// CountPostReactionsByPostIDs loads post reaction totals for all visible posts.
func CountPostReactionsByPostIDs(db *sql.DB, postIDs []int) (map[int]ReactionCounts, error) {
	counts := make(map[int]ReactionCounts, len(postIDs))
	if len(postIDs) == 0 {
		return counts, nil
	}

	query := fmt.Sprintf(`
		SELECT post_id, reaction_type, COUNT(*)
		FROM post_reactions
		WHERE post_id IN (%s)
		GROUP BY post_id, reaction_type
	`, placeholders(len(postIDs)))
	rows, err := db.Query(query, intsToArgs(postIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID int
		var reactionType string
		var count int
		if err := rows.Scan(&postID, &reactionType, &count); err != nil {
			return nil, err
		}
		reactionCounts := counts[postID]
		if reactionType == "like" {
			reactionCounts.Likes = count
		}
		if reactionType == "dislike" {
			reactionCounts.Dislikes = count
		}
		counts[postID] = reactionCounts
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}

// GetCategoriesByPostIDs loads all category labels for the visible posts.
func GetCategoriesByPostIDs(db *sql.DB, postIDs []int) (map[int][]models.Category, error) {
	categoriesByPost := make(map[int][]models.Category, len(postIDs))
	if len(postIDs) == 0 {
		return categoriesByPost, nil
	}

	query := fmt.Sprintf(`
		SELECT pc.post_id, c.id, c.name
		FROM post_categories pc
		JOIN categories c ON c.id = pc.category_id
		WHERE pc.post_id IN (%s)
		ORDER BY pc.post_id ASC, c.name ASC
	`, placeholders(len(postIDs)))
	rows, err := db.Query(query, intsToArgs(postIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var postID int
		var category models.Category
		if err := rows.Scan(&postID, &category.ID, &category.Name); err != nil {
			return nil, err
		}
		categoriesByPost[postID] = append(categoriesByPost[postID], category)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categoriesByPost, nil
}

// GetRecentCommentsByPostIDs loads at most limitPerPost newest comments per visible post.
func GetRecentCommentsByPostIDs(db *sql.DB, postIDs []int, limitPerPost int) (map[int][]models.Comment, error) {
	commentsByPost := make(map[int][]models.Comment, len(postIDs))
	if len(postIDs) == 0 || limitPerPost <= 0 {
		return commentsByPost, nil
	}

	query := `
		SELECT id, author_id, post_id, content, created_at
		FROM comments
		WHERE post_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`
	for _, postID := range postIDs {
		rows, err := db.Query(query, postID, limitPerPost)
		if err != nil {
			return nil, err
		}

		var newestFirst []models.Comment
		for rows.Next() {
			var comment models.Comment
			if err := rows.Scan(&comment.ID, &comment.AuthorID, &comment.PostID, &comment.Content, &comment.CreatedAt); err != nil {
				rows.Close()
				return nil, err
			}
			newestFirst = append(newestFirst, comment)
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()

		for i := len(newestFirst) - 1; i >= 0; i-- {
			commentsByPost[postID] = append(commentsByPost[postID], newestFirst[i])
		}
	}

	return commentsByPost, nil
}

// CountCommentReactionsByCommentIDs loads comment reaction totals for visible comments.
func CountCommentReactionsByCommentIDs(db *sql.DB, commentIDs []int) (map[int]ReactionCounts, error) {
	counts := make(map[int]ReactionCounts, len(commentIDs))
	if len(commentIDs) == 0 {
		return counts, nil
	}

	query := fmt.Sprintf(`
		SELECT comment_id, reaction_type, COUNT(*)
		FROM comment_reactions
		WHERE comment_id IN (%s)
		GROUP BY comment_id, reaction_type
	`, placeholders(len(commentIDs)))
	rows, err := db.Query(query, intsToArgs(commentIDs)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var commentID int
		var reactionType string
		var count int
		if err := rows.Scan(&commentID, &reactionType, &count); err != nil {
			return nil, err
		}
		reactionCounts := counts[commentID]
		if reactionType == "like" {
			reactionCounts.Likes = count
		}
		if reactionType == "dislike" {
			reactionCounts.Dislikes = count
		}
		counts[commentID] = reactionCounts
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return counts, nil
}
