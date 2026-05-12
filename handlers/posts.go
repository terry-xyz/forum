package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

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

		// GET renders the post creation form and its category checkboxes.
		if r.Method == http.MethodGet {
			// Categories come from the database so form options stay in sync with
			// the schema seed data.
			categories, err := database.GetAllCategories(db)
			if err != nil {
				http.Error(w, "unable to fetch categories", http.StatusInternalServerError)
				return
			}

			// The handler writes simple HTML directly instead of using templates.
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/create-post">
				Title: <input name="title"><br>
				Content:
				<textarea name="content"></textarea><br>
				Categories:<br>
			`))
			// Each category becomes one checkbox sharing the category_ids field name.
			for _, c := range categories {
				w.Write([]byte(`
				<label>
					<input type="checkbox" name="category_ids" value="` + strconv.Itoa(c.ID) + `">` + c.Name +
					`</label><br>`,
				))
			}
			// Close the form after all dynamic checkbox rows have been written.
			w.Write([]byte(`
				<button type="submit">Create Post</button>
			</form>
			`))
			return
		}

		// POST validates submitted data, creates the post, and links categories.
		if r.Method == http.MethodPost {
			// ParseForm is required here because category_ids is read from r.Form
			// as a repeated value.
			err := r.ParseForm()
			if err != nil {
				http.Error(w, "bad form data", http.StatusBadRequest)
				return
			}
			// Scalar fields come from the submitted form body.
			title := r.FormValue("title")
			content := r.FormValue("content")

			// Convert the repeated checkbox values from strings into database IDs.
			categoryIDStrs := r.Form["category_ids"]
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
				http.Error(w, "select at least one category", http.StatusBadRequest)
				return
			}

			// Insert the post first so its generated ID can be used in post_categories.
			postID, err := database.CreatePost(db, id, title, content)
			if err != nil {
				http.Error(w, "unable to create post", http.StatusInternalServerError)
				return
			}

			// Add the selected category relationships after the post exists.
			err = database.AddCategoriesToPost(db, postID, categoryIDs)
			if err != nil {
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

		// Query by author_id so the page shows only posts owned by this user.
		posts, err := database.GetPostsByAuthorID(db, userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if len(posts) == 0 {
			// Empty results are valid for new users, so render an empty-state message.
			w.Write([]byte("<p>You have not created any posts yet.</p>"))
			return
		} else {
			// renderPosts needs the current user to decide whether reaction and
			// comment forms should be shown.
			currentUser, err := database.GetUserByID(db, userID)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if currentUser == nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			// Reuse the feed renderer so reaction and comment markup stays identical.
			err = renderPosts(w, db, posts, currentUser)
			if err != nil {
				http.Error(w, "failed to render posts", http.StatusInternalServerError)
				return
			}
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

		// Join through post_reactions so only posts this user liked are returned.
		posts, err := database.GetLikedPostsByUserID(db, userID)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		if len(posts) == 0 {
			// Empty results are valid when the user has not liked anything yet.
			w.Write([]byte("<p>You don't have any liked posts yet.</p>"))
			return
		} else {
			// renderPosts needs the current user to decide whether reaction and
			// comment forms should be shown.
			currentUser, err := database.GetUserByID(db, userID)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			if currentUser == nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			// Reuse the feed renderer so reaction and comment markup stays identical.
			err = renderPosts(w, db, posts, currentUser)
			if err != nil {
				http.Error(w, "failed to render posts", http.StatusInternalServerError)
				return
			}
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
