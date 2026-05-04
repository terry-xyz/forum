package handlers

import (
	"database/sql"
	"forum/database"
	"net/http"
	"strconv"
)

// HomeHandler renders the forum home page for users with a valid session cookie.
func HomeHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			// Session cookies currently store the user ID, so reject values that cannot be parsed.
			cookieID, err := strconv.Atoi(cookie.Value)
			if err != nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			user, err := database.GetUserByID(db, cookieID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
			}

			posts, err := database.GetAllPosts(db)
			if err != nil {
				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}
			for _, p := range posts {
				user, err := database.GetUserByID(db, p.AuthorID)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				if user == nil {
					http.Error(w, "invalid post author", http.StatusInternalServerError)
					return
				}
				w.Write([]byte(
					`<h3>` + p.Title + `</h3>` +
						`<p>` + p.Content + `</p>` +
						`<small>Author: ` + user.Username + `</small><hr>`,
				))
			}

			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
