package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

func renderCommentValidationError(w http.ResponseWriter, db *sql.DB, currentUser *models.User, csrfToken string, postID int, message string) {
	posts, err := database.GetAllPostsPage(db, feedPageSize+1, 0)
	if err != nil {
		renderHTTPError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	posts, hasNext := trimPage(posts)

	pagination := paginationView{
		HasNext: hasNext,
		NextURL: pageURL("/", nil, 2),
	}
	if err := renderHomePage(w, db, posts, currentUser, "", csrfToken, pagination, postID, message); err != nil {
		renderHTTPError(w, http.StatusInternalServerError, "failed to render home page")
	}
}

// CreateCommentHandler creates a comment for the logged-in user.
func CreateCommentHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database dependency for comment writes.
	return func(w http.ResponseWriter, r *http.Request) {
		// Comments are submitted from forms, so this endpoint only accepts POST.
		if r.Method != http.MethodPost {
			renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// Require a session before accepting comment content.
		cookie, err := r.Cookie("session")
		if err != nil {
			renderHTTPError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		// Convert the session value into the comment author ID.
		authorID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if authorID == 0 {
			renderHTTPError(w, http.StatusUnauthorized, "invalid session")
			return
		}
		currentUser, err := database.GetUserByID(db, authorID)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
			return
		}
		csrfToken, err := database.GetCSRFTokenBySessionID(db, cookie.Value)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		// The form posts the owning post ID as a hidden field.
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
		// Empty comments are rejected before hitting the database.
		content := r.FormValue("content")
		if strings.TrimSpace(content) == "" {
			renderCommentValidationError(w, db, currentUser, csrfToken, postID, "comment cannot be empty")
			return
		}
		if exceedsCharacterLimit(content, maxCommentContentChars) {
			renderCommentValidationError(w, db, currentUser, csrfToken, postID, "comment cannot exceed 280 characters")
			return
		}

		// Persist the comment with the session user as author.
		err = database.CreateComment(db, authorID, postID, content)
		if err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to create comment")
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
			renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		// The session is needed to prevent deleting another user's comment.
		cookie, err := r.Cookie("session")
		if err != nil {
			renderHTTPError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		// Resolve the session token into the author ID used by the delete query.
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

		// The comment ID comes from the hidden field beside the rendered comment.
		commentIDStr := r.FormValue("comment_id")
		commentID, err := strconv.Atoi(commentIDStr)
		if err != nil {
			renderHTTPError(w, http.StatusBadRequest, "invalid comment id")
			return
		}

		// The database helper includes author_id in the WHERE clause so users
		// cannot remove comments they do not own.
		err = database.DeleteCommentByIDAndAuthorID(db, commentID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				renderHTTPError(w, http.StatusForbidden, "comment not found or not yours")
				return
			}

			renderHTTPError(w, http.StatusInternalServerError, "unable to delete comment")
			return
		}

		// Return to the feed after the comment has been removed.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
