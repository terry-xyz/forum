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
		// The session cookie currently stores the numeric user ID directly.
		id, err := strconv.Atoi(cookie.Value)
		if err != nil {
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
