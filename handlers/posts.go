package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"forum/models"
	"html"
	"net/http"
	"strconv"
	"strings"
)

func createPostFormHTML(csrfToken string, categories []models.Category, title string, content string, selectedCategoryIDs []string) string {
	selected := make(map[string]bool, len(selectedCategoryIDs))
	for _, categoryID := range selectedCategoryIDs {
		selected[categoryID] = true
	}

	var form strings.Builder
	form.WriteString(`
			<form method="POST" action="/create-post">
				<input type="hidden" name="csrf_token" value="` + html.EscapeString(csrfToken) + `">
				<label>Title <input name="title" value="` + html.EscapeString(title) + `"></label>
				<label>Content <textarea name="content">` + html.EscapeString(content) + `</textarea></label>
				<fieldset>
					<legend>Categories</legend>
			`)
	for _, c := range categories {
		categoryID := strconv.Itoa(c.ID)
		checked := ""
		if selected[categoryID] {
			checked = " checked"
		}
		form.WriteString(`
				<label>
					<input type="checkbox" name="category_ids" value="` + categoryID + `"` + checked + `>` + html.EscapeString(c.Name) +
			`</label>`)
	}
	form.WriteString(`
				</fieldset>
				<button type="submit">Create Post</button>
			</form>
			`)

	return form.String()
}

func renderCreatePostFormPage(w http.ResponseWriter, db *sql.DB, csrfToken string, alert string, title string, content string, selectedCategoryIDs []string) bool {
	// Categories come from the database so form options stay in sync with
	// the schema seed data.
	categories, err := database.GetAllCategories(db)
	if err != nil {
		http.Error(w, "unable to fetch categories", http.StatusInternalServerError)
		return false
	}

	renderFormPage(w, "Create post", alert, createPostFormHTML(csrfToken, categories, title, content, selectedCategoryIDs))
	return true
}

// CreatePostHandler serves the post form and creates posts for the logged-in user.
func CreatePostHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle once when the route is registered.
	return func(w http.ResponseWriter, r *http.Request) {

		// Creating or viewing the post form requires a session.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		// Resolve the opaque session token into the post author ID.
		id, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if id == 0 {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}
		csrfToken, err := database.GetCSRFTokenBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if csrfToken == "" {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		// GET renders the post creation form and its category checkboxes.
		if r.Method == http.MethodGet {
			renderCreatePostFormPage(w, db, csrfToken, "", "", "", nil)
			return
		}

		// POST validates submitted data, creates the post, and links categories.
		if r.Method == http.MethodPost {
			if !parseLimitedForm(w, r) {
				return
			}
			if !validCSRFToken(db, r, cookie.Value) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			// Scalar fields come from the submitted form body.
			rawTitle := r.FormValue("title")
			rawContent := r.FormValue("content")
			title := strings.TrimSpace(rawTitle)
			content := strings.TrimSpace(rawContent)
			categoryIDStrs := r.Form["category_ids"]

			if title == "" {
				renderCreatePostFormPage(w, db, csrfToken, "title cannot be empty", rawTitle, rawContent, categoryIDStrs)
				return
			}
			if content == "" {
				renderCreatePostFormPage(w, db, csrfToken, "content cannot be empty", rawTitle, rawContent, categoryIDStrs)
				return
			}
			if exceedsCharacterLimit(title, maxPostTitleChars) {
				renderCreatePostFormPage(w, db, csrfToken, "title cannot exceed 280 characters", rawTitle, rawContent, categoryIDStrs)
				return
			}
			if exceedsCharacterLimit(content, maxPostContentChars) {
				renderCreatePostFormPage(w, db, csrfToken, "content cannot exceed 280 characters", rawTitle, rawContent, categoryIDStrs)
				return
			}

			// Convert the repeated checkbox values from strings into database IDs.
			var categoryIDs []int
			for _, c := range categoryIDStrs {
				categoryID, err := strconv.Atoi(c)
				if err != nil {
					http.Error(w, "invalid category id", http.StatusBadRequest)
					return
				}

				categoryIDs = append(categoryIDs, categoryID)
			}

			// A post must belong to at least one category for filtering to work.
			if len(categoryIDs) == 0 {
				renderCreatePostFormPage(w, db, csrfToken, "select at least one category", rawTitle, rawContent, categoryIDStrs)
				return
			}

			// Insert the post and its category links together so failed category
			// validation cannot leave an uncategorized post behind.
			_, err := database.CreatePostWithCategories(db, id, title, content, categoryIDs)
			if err != nil {
				if errors.Is(err, database.ErrInvalidCategoryID) {
					http.Error(w, "invalid category id", http.StatusBadRequest)
					return
				}
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			// Redirect to the feed after creation to avoid duplicate form submits.
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Only GET and POST make sense for the post creation endpoint.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

	}
}

// MyPostsHandler renders posts created by the current session user.
func MyPostsHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database dependency once so each request can reuse it.
	return func(w http.ResponseWriter, r *http.Request) {
		// The filtered list is read-only, so non-GET methods are rejected early.
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// The session cookie identifies which author's posts should be listed.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Resolve the opaque session token before using it in the author filter.
		userID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if userID == 0 {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		page, err := parsePage(r)
		if err != nil {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		offset := (page - 1) * feedPageSize

		// Query by author_id so the page shows only posts owned by this user.
		posts, err := database.GetPostsByAuthorIDPage(db, userID, feedPageSize+1, offset)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		posts, hasNext := trimPage(posts)

		currentUser, err := database.GetUserByID(db, userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if currentUser == nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		emptyMessage := ""
		if len(posts) == 0 {
			// Empty results are valid for new users, so render an empty-state message.
			emptyMessage = "You have not created any posts yet."
		}
		csrfToken, err := database.GetCSRFTokenBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if csrfToken == "" {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		if err := renderHomePage(w, db, posts, currentUser, emptyMessage, csrfToken, paginationForRequest(r, page, hasNext)); err != nil {
			http.Error(w, "failed to render posts", http.StatusInternalServerError)
			return
		}
	}
}

// LikedPostsHandler renders posts liked by the current session user.
func LikedPostsHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database dependency once so every request shares it.
	return func(w http.ResponseWriter, r *http.Request) {
		// The liked-posts page only reads data, so it accepts GET requests only.
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// The session cookie identifies which user's reaction rows should be used.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Invalid cookie values cannot safely join against reaction user IDs.
		userID, err := database.GetUserIDBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if userID == 0 {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		page, err := parsePage(r)
		if err != nil {
			http.Error(w, "invalid page", http.StatusBadRequest)
			return
		}
		offset := (page - 1) * feedPageSize

		// Join through post_reactions so only posts this user liked are returned.
		posts, err := database.GetLikedPostsByUserIDPage(db, userID, feedPageSize+1, offset)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		posts, hasNext := trimPage(posts)

		currentUser, err := database.GetUserByID(db, userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if currentUser == nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		emptyMessage := ""
		if len(posts) == 0 {
			// Empty results are valid when the user has not liked anything yet.
			emptyMessage = "You don't have any liked posts yet."
		}
		csrfToken, err := database.GetCSRFTokenBySessionID(db, cookie.Value)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		if csrfToken == "" {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		if err := renderHomePage(w, db, posts, currentUser, emptyMessage, csrfToken, paginationForRequest(r, page, hasNext)); err != nil {
			http.Error(w, "failed to render posts", http.StatusInternalServerError)
			return
		}
	}
}

// DeletePostHandler removes a post only when the current session user owns it.
func DeletePostHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle so the delete path can verify ownership.
	return func(w http.ResponseWriter, r *http.Request) {
		// Deletes come from rendered forms, so this endpoint only accepts POST.
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// A session is required because the delete query checks the post author.
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Resolve the session into the user ID used by the ownership check.
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

		// The post ID comes from the hidden field rendered beside owned posts.
		postIDStr := r.FormValue("post_id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "invalid post id", http.StatusBadRequest)
			return
		}

		// The database helper deletes related rows in one transaction after
		// confirming this user is the author.
		err = database.DeletePostByIDAndAuthorID(db, postID, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "post not found or not yours", http.StatusForbidden)
				return
			}

			http.Error(w, "unable to delete post", http.StatusInternalServerError)
			return
		}

		// Return to the feed where the removed post should no longer appear.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
