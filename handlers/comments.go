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
		authorID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if authorID == 0 {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}
		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
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
		if exceedsCharacterLimit(content, maxCommentContentChars) {
			http.Error(w, "comment cannot exceed 280 characters", http.StatusBadRequest)
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

// DeleteCommentHandler removes a comment only when the current session user owns it.
func DeleteCommentHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle for the ownership-scoped delete.
	return func(w http.ResponseWriter, r *http.Request) {
		// Comments are deleted from rendered forms, so only POST is accepted.
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// The session is needed to prevent deleting another user's comment.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Resolve the session token into the author ID used by the delete query.
		userID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if userID == 0 {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}
		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		// The comment ID comes from the hidden field beside the rendered comment.
		commentIDStr := r.FormValue("comment_id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "invalid comment id", http.StatusBadRequest)
			return
		}

		// The database helper includes author_id in the WHERE clause so users
		// cannot remove comments they do not own.
		err = database.DeleteCommentByIDAndAuthorID(db, commentID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "comment not found or not yours", http.StatusForbidden)
				return
			}

			http.Error(w, "unable to delete comment", http.StatusInternalServerError)
			return
		}

		// Return to the feed after the comment has been removed.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
