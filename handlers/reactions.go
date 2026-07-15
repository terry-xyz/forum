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
			renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// Reactions belong to the logged-in user.
		cookie, err := r.Cookie("session")
		if err != nil {
			renderHTTPError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Convert the session cookie into the reacting user's ID.
		userID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if userID == 0 {
			renderHTTPError(w, http.StatusUnauthorized, "invalid session")
			return
		}
		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
			return
		}

		// The post ID arrives as a hidden form field.
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			renderHTTPError(w, http.StatusBadRequest, "invalid post id")
			return
		}
		postExists, err := database.PostExists(db, postID)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to validate post")
			return
		}
		if !postExists {
			renderHTTPError(w, http.StatusNotFound, "post not found")
			return
		}
		// Only the two schema-supported reaction values are allowed through.
		reactionType := r.FormValue("reaction_type")
		if reactionType != "like" && reactionType != "dislike" {
			renderHTTPError(w, http.StatusBadRequest, "invalid reaction type")
			return
		}

		// The database upsert handles both first reactions and reaction changes.
		err = database.ReactToPost(db, userID, postID, reactionType)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to react to post")
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
			renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// A reaction must be tied to a logged-in user.
		cookie, err := r.Cookie("session")
		if err != nil {
			renderHTTPError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Convert the current session into the reacting user ID.
		userID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if userID == 0 {
			renderHTTPError(w, http.StatusUnauthorized, "invalid session")
			return
		}
		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
			return
		}

		// The comment ID arrives from the hidden field beside each rendered comment.
		commentIDStr := r.FormValue("comment_id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			renderHTTPError(w, http.StatusBadRequest, "invalid comment id")
			return
		}
		commentExists, err := database.CommentExists(db, commentID)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to validate comment")
			return
		}
		if !commentExists {
			renderHTTPError(w, http.StatusNotFound, "comment not found")
			return
		}
		// Reject anything outside the two allowed reaction values before SQL.
		reactionType := r.FormValue("reaction_type")
		if reactionType != "like" && reactionType != "dislike" {
			renderHTTPError(w, http.StatusBadRequest, "invalid reaction type")
			return
		}

		// Upsert the reaction so a user can switch between like and dislike.
		err = database.ReactToComment(db, userID, commentID, reactionType)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to react to comment")
			return
		}

		// Return to the feed where the changed reaction totals are rendered.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
