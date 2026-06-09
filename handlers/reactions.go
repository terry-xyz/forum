package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

// ReactPostHandler records a like or dislike for a post.
func ReactPostHandler(db *sql.DB) http.HandlerFunc {

	// Capture the shared database handle for reaction writes.
	return func(w http.ResponseWriter, r *http.Request) {
		// Reactions are submitted from forms, so only POST is accepted.
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Reactions belong to the logged-in user.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert the session cookie into the reacting user's ID.
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

		// The post ID arrives as a hidden form field.
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}
		// Only the two schema-supported reaction values are allowed through.
		reactionType := r.FormValue("reaction_type")
		if reactionType != "like" && reactionType != "dislike" {
			http.Error(w, "invalid reaction type", http.StatusBadRequest)
			return
		}

		// The database upsert handles both first reactions and reaction changes.
		err = database.ReactToPost(db, userID, postID, reactionType)
		if err != nil {
			http.Error(w, "unable to react to post", http.StatusInternalServerError)
			return
		}

		// Reload the feed so the updated counts are visible.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

// ReactCommentHandler records a like or dislike for a comment.
func ReactCommentHandler(db *sql.DB) http.HandlerFunc {

	// Capture the shared database handle for comment reaction writes.
	return func(w http.ResponseWriter, r *http.Request) {
		// Comment reactions are form submissions, so this endpoint is POST-only.
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// A reaction must be tied to a logged-in user.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Convert the current session into the reacting user ID.
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

		// The comment ID arrives from the hidden field beside each rendered comment.
		commentIDStr := r.FormValue("comment_id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			http.Error(w, "invalid comment id", http.StatusBadRequest)
			return
		}
		// Reject anything outside the two allowed reaction values before SQL.
		reactionType := r.FormValue("reaction_type")
		if reactionType != "like" && reactionType != "dislike" {
			http.Error(w, "invalid reaction type", http.StatusBadRequest)
			return
		}

		// Upsert the reaction so a user can switch between like and dislike.
		err = database.ReactToComment(db, userID, commentID, reactionType)
		if err != nil {
			http.Error(w, "unable to react to comment", http.StatusInternalServerError)
			return
		}

		// Return to the feed where the changed reaction totals are rendered.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
