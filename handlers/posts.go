package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

// CreatePostHandler serves the post form and creates posts for the logged-in user.
func CreatePostHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		id, err := strconv.Atoi(cookie.Value)
		if err != nil {
			http.Error(w, "invalid session", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/create-post">
				Title: <input name="title"><br>
				Content: <textarea name="content"></textarea><br>
				<button type="submit">Create Post</button>
			</form>
			`))
			return
		}

		if r.Method == http.MethodPost {
			title := r.FormValue("title")
			content := r.FormValue("content")

			err := database.CreatePost(db, id, title, content)
			if err != nil {
				http.Error(w, "unable to create post", http.StatusInternalServerError)
				return
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)

	}
}
