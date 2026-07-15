package database

import (
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

const seedPassword = "password123"

// seedUser carries the stable account fields used by SeedFakeData.
type seedUser struct {
	Email    string
	Username string
}

// seedPost stores post fixtures with category names so seed data can resolve IDs at insert time.
type seedPost struct {
	AuthorEmail string
	Title       string
	Content     string
	Categories  []string
}

// seedComment stores comment fixtures by natural keys so the seed can run idempotently.
type seedComment struct {
	AuthorEmail string
	PostTitle   string
	Content     string
}

// SeedFakeData inserts a small deterministic set of users, posts, comments, and reactions.
func SeedFakeData(db *sql.DB) error {
	// Keep fixtures in code so the seed command is deterministic and easy to reset.
	users := []seedUser{
		{Email: "alice@example.com", Username: "alice"},
		{Email: "ben@example.com", Username: "ben"},
		{Email: "cora@example.com", Username: "cora"},
	}

	posts := []seedPost{
		{
			AuthorEmail: "alice@example.com",
			Title:       "Welcome to the forum",
			Content:     "Say hello and share what you are working on.",
			Categories:  []string{"General"},
		},
		{
			AuthorEmail: "ben@example.com",
			Title:       "Favorite Go packages",
			Content:     "What packages have made your Go projects easier to build?",
			Categories:  []string{"Programming"},
		},
		{
			AuthorEmail: "cora@example.com",
			Title:       "Weekend game recommendations",
			Content:     "Looking for a short game to play over the weekend.",
			Categories:  []string{"Gaming", "General"},
		},
		{
			AuthorEmail: "alice@example.com",
			Title:       "Albums for coding sessions",
			Content:     "Drop music that helps you focus while building things.",
			Categories:  []string{"Music", "Programming"},
		},
	}

	comments := []seedComment{
		{AuthorEmail: "ben@example.com", PostTitle: "Welcome to the forum", Content: "Glad this is up and running."},
		{AuthorEmail: "cora@example.com", PostTitle: "Welcome to the forum", Content: "Hello everyone."},
		{AuthorEmail: "alice@example.com", PostTitle: "Favorite Go packages", Content: "sqlite3 and bcrypt are already useful here."},
		{AuthorEmail: "cora@example.com", PostTitle: "Favorite Go packages", Content: "I like keeping dependencies small."},
		{AuthorEmail: "alice@example.com", PostTitle: "Weekend game recommendations", Content: "Try something co-op if you have time."},
		{AuthorEmail: "ben@example.com", PostTitle: "Albums for coding sessions", Content: "Instrumental playlists work best for me."},
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// Rollback is deferred so any early return leaves the database unchanged;
	// after Commit succeeds, Rollback becomes a harmless no-op.
	defer tx.Rollback()

	for _, user := range users {
		if _, err := ensureSeedUser(tx, user); err != nil {
			return err
		}
	}

	for _, post := range posts {
		postID, err := ensureSeedPost(tx, post)
		if err != nil {
			return err
		}
		for _, category := range post.Categories {
			if err := ensureSeedPostCategory(tx, postID, category); err != nil {
				return err
			}
		}
	}

	for _, comment := range comments {
		if _, err := ensureSeedComment(tx, comment); err != nil {
			return err
		}
	}

	if err := seedReactions(tx); err != nil {
		return err
	}

	return tx.Commit()
}

// ensureSeedUser returns an existing seed user ID or inserts the user with a hashed demo password.
func ensureSeedUser(tx *sql.Tx, user seedUser) (int, error) {
	// Email is unique in the schema, so it is the stable key for rerunnable seeds.
	id, err := seedUserIDByEmail(tx, user.Email)
	if err != nil || id != 0 {
		return id, err
	}

	// Seeded accounts use bcrypt too so login behavior matches registered users.
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		user.Email,
		user.Username,
		string(hashedPassword),
	)
	if err != nil {
		return 0, err
	}

	insertedID, err := result.LastInsertId()
	return int(insertedID), err
}

// ensureSeedPost returns an existing seed post ID or inserts it for the configured author.
func ensureSeedPost(tx *sql.Tx, post seedPost) (int, error) {
	// Titles are unique within this fixture set, so they make the seed rerunnable.
	id, err := seedPostIDByTitle(tx, post.Title)
	if err != nil || id != 0 {
		return id, err
	}

	// The author fixture must already exist because users are seeded before posts.
	authorID, err := seedUserIDByEmail(tx, post.AuthorEmail)
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(
		"INSERT INTO posts (author_id, title, content) VALUES (?, ?, ?)",
		authorID,
		post.Title,
		post.Content,
	)
	if err != nil {
		return 0, err
	}

	insertedID, err := result.LastInsertId()
	return int(insertedID), err
}

// ensureSeedComment returns an existing seed comment ID or inserts it below its configured post.
func ensureSeedComment(tx *sql.Tx, comment seedComment) (int, error) {
	// Post title plus comment content identifies fixture comments across seed runs.
	id, err := seedCommentID(tx, comment)
	if err != nil || id != 0 {
		return id, err
	}

	// The author and post fixtures must already exist because they are seeded first.
	authorID, err := seedUserIDByEmail(tx, comment.AuthorEmail)
	if err != nil {
		return 0, err
	}

	postID, err := seedPostIDByTitle(tx, comment.PostTitle)
	if err != nil {
		return 0, err
	}

	result, err := tx.Exec(
		"INSERT INTO comments (author_id, post_id, content) VALUES (?, ?, ?)",
		authorID,
		postID,
		comment.Content,
	)
	if err != nil {
		return 0, err
	}

	insertedID, err := result.LastInsertId()
	return int(insertedID), err
}

// ensureSeedPostCategory links a seed post to a category, creating the category if needed.
func ensureSeedPostCategory(tx *sql.Tx, postID int, categoryName string) error {
	// Resolve by category name so fixture data does not depend on numeric IDs.
	categoryID, err := seedCategoryIDByName(tx, categoryName)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		"INSERT OR IGNORE INTO post_categories (post_id, category_id) VALUES (?, ?)",
		postID,
		categoryID,
	)
	return err
}

// seedReactions upserts deterministic post and comment reactions for the seed fixtures.
func seedReactions(tx *sql.Tx) error {
	// Natural keys keep the reaction fixtures readable while helper lookups supply IDs.
	postReactions := []struct {
		UserEmail    string
		PostTitle    string
		ReactionType string
	}{
		{UserEmail: "alice@example.com", PostTitle: "Favorite Go packages", ReactionType: "like"},
		{UserEmail: "alice@example.com", PostTitle: "Weekend game recommendations", ReactionType: "like"},
		{UserEmail: "ben@example.com", PostTitle: "Welcome to the forum", ReactionType: "like"},
		{UserEmail: "ben@example.com", PostTitle: "Albums for coding sessions", ReactionType: "dislike"},
		{UserEmail: "cora@example.com", PostTitle: "Welcome to the forum", ReactionType: "like"},
		{UserEmail: "cora@example.com", PostTitle: "Favorite Go packages", ReactionType: "like"},
	}

	for _, reaction := range postReactions {
		// Resolve user and post IDs inside the transaction so reactions point at
		// the rows created or reused earlier in this seed run.
		userID, err := seedUserIDByEmail(tx, reaction.UserEmail)
		if err != nil {
			return err
		}
		postID, err := seedPostIDByTitle(tx, reaction.PostTitle)
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO post_reactions (user_id, post_id, reaction_type)
			VALUES (?, ?, ?)
			ON CONFLICT(user_id, post_id)
			DO UPDATE SET reaction_type = excluded.reaction_type
		`, userID, postID, reaction.ReactionType)
		if err != nil {
			return err
		}
	}

	// Comment reactions use post title and comment content because comments do
	// not have a human-readable unique field by themselves.
	commentReactions := []struct {
		UserEmail      string
		CommentPost    string
		CommentContent string
		ReactionType   string
	}{
		{UserEmail: "alice@example.com", CommentPost: "Welcome to the forum", CommentContent: "Glad this is up and running.", ReactionType: "like"},
		{UserEmail: "alice@example.com", CommentPost: "Favorite Go packages", CommentContent: "I like keeping dependencies small.", ReactionType: "like"},
		{UserEmail: "ben@example.com", CommentPost: "Welcome to the forum", CommentContent: "Hello everyone.", ReactionType: "like"},
		{UserEmail: "ben@example.com", CommentPost: "Weekend game recommendations", CommentContent: "Try something co-op if you have time.", ReactionType: "like"},
		{UserEmail: "cora@example.com", CommentPost: "Favorite Go packages", CommentContent: "sqlite3 and bcrypt are already useful here.", ReactionType: "like"},
		{UserEmail: "cora@example.com", CommentPost: "Albums for coding sessions", CommentContent: "Instrumental playlists work best for me.", ReactionType: "dislike"},
	}

	for _, reaction := range commentReactions {
		// Resolve comment IDs from fixture keys before upserting the reaction row.
		userID, err := seedUserIDByEmail(tx, reaction.UserEmail)
		if err != nil {
			return err
		}
		commentID, err := seedCommentID(tx, seedComment{
			PostTitle: reaction.CommentPost,
			Content:   reaction.CommentContent,
		})
		if err != nil {
			return err
		}

		_, err = tx.Exec(`
			INSERT INTO comment_reactions (user_id, comment_id, reaction_type)
			VALUES (?, ?, ?)
			ON CONFLICT(user_id, comment_id)
			DO UPDATE SET reaction_type = excluded.reaction_type
		`, userID, commentID, reaction.ReactionType)
		if err != nil {
			return err
		}
	}

	return nil
}

// seedUserIDByEmail returns the user ID for a seed email, or zero when it is missing.
func seedUserIDByEmail(tx *sql.Tx, email string) (int, error) {
	var id int
	err := tx.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// seedPostIDByTitle returns the post ID for a seed title, or zero when it is missing.
func seedPostIDByTitle(tx *sql.Tx, title string) (int, error) {
	var id int
	err := tx.QueryRow("SELECT id FROM posts WHERE title = ?", title).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// seedCommentID returns the comment ID for a seed post/content pair, or zero when missing.
func seedCommentID(tx *sql.Tx, comment seedComment) (int, error) {
	// Resolve the post first because comments are unique only within a post in this fixture set.
	postID, err := seedPostIDByTitle(tx, comment.PostTitle)
	if err != nil {
		return 0, err
	}

	var id int
	err = tx.QueryRow(
		"SELECT id FROM comments WHERE post_id = ? AND content = ?",
		postID,
		comment.Content,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

// seedCategoryIDByName returns an existing category ID or inserts the category for seed data.
func seedCategoryIDByName(tx *sql.Tx, name string) (int, error) {
	var id int
	err := tx.QueryRow("SELECT id FROM categories WHERE name = ?", name).Scan(&id)
	if err == sql.ErrNoRows {
		// Seed data can introduce categories beyond schema defaults, so create
		// missing names instead of treating them as broken fixtures.
		result, insertErr := tx.Exec("INSERT INTO categories (name) VALUES (?)", name)
		if insertErr != nil {
			return 0, insertErr
		}
		insertedID, lastIDErr := result.LastInsertId()
		return int(insertedID), lastIDErr
	}
	return id, err
}
