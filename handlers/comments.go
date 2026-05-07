package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

// CreateCommentHandler creates a comment for the logged-in user.
func CreateCommentHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database dependency for comment writes.
	return func(w http.ResponseWriter, r *http.Request) {
		// Comments are submitted from forms, so this endpoint only accepts POST.
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Require a session before accepting comment content.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// Convert the session value into the comment author ID.
		authorID, err := strconv.Atoi(cookie.Value)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		// The form posts the owning post ID as a hidden field.
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}
		// Empty comments are rejected before hitting the database.
		content := r.FormValue("content")
		if content == "" {
			http.Error(w, "comment cannot be empty", http.StatusBadRequest)
			return
		}

		// Persist the comment with the session user as author.
		err = database.CreateComment(db, authorID, postID, content)
		if err != nil {
			http.Error(w, "unable to create comment", http.StatusInternalServerError)
			return
		}

		// Return to the feed where the new comment should appear.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
