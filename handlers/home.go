package handlers

import (
	"database/sql"
	"fmt"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
)

// HomeHandler renders the forum home page for guests and adds user actions when a session is valid.
func HomeHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle for all requests served by this route.
	return func(w http.ResponseWriter, r *http.Request) {

		// The home page is currently read-only, so only GET is supported.
		if r.Method == http.MethodGet {

			var currentUser *models.User

			// Guests can view the feed, so a missing or invalid session simply
			// leaves currentUser nil and hides authenticated actions.
			cookie, err := r.Cookie("session")
			if err == nil {
				userID, err := database.GetUserIDBySessionID(db, cookie.Value)
				if err != nil {
					http.Error(w, "internal error", http.StatusInternalServerError)
					return
				}
				if userID != 0 {
					currentUser, err = database.GetUserByID(db, userID)
					if err != nil {
						http.Error(w, "internal error", http.StatusInternalServerError)
						return
					}
				}
			}

			// Choose between the default feed and a category-filtered feed based
			// on the optional query string.
			var posts []models.Post
			categoryIDStr := r.URL.Query().Get("category_id")
			if categoryIDStr == "" {
				// No filter means the full post list should be shown.
				posts, err = database.GetAllPosts(db)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			} else {
				// Bad query values are client errors because category_id is part of
				// the request URL.
				categoryID, err := strconv.Atoi(categoryIDStr)
				if err != nil {
					http.Error(w, "invalid category id", http.StatusBadRequest)
					return
				}
				// A valid category ID narrows the feed to posts linked to it.
				posts, err = database.GetPostsByCategoryID(db, categoryID)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
			}

			// All response bodies from this handler are hand-written HTML.
			w.Header().Set("Content-Type", "text/html")

			// Load category filters after post selection so the filter bar always
			// contains every available category.
			allCategories, err := database.GetAllCategories(db)
			if err != nil {
				http.Error(w, "failed to load category filters", http.StatusInternalServerError)
				return
			}
			// Render a simple filter bar with an "All" link plus one link per category.
			w.Write([]byte(`<p><a href="/">All</a> `))
			for _, c := range allCategories {
				w.Write([]byte(
					`<a href="/?category_id=` + strconv.Itoa(c.ID) + `">` + c.Name + `</a> `,
				))
			}
			w.Write([]byte(`</p><hr>`))

			if currentUser != nil {
				w.Write([]byte(`
					<a href="/create-post">Create post</a>
					<a href="/my-posts">My posts</a>
					<a href="/liked-posts">Liked posts</a>
					<a href="/logout">Logout</a>
				`))
			} else {
				w.Write([]byte(`
					<a href="/login">Login</a>
					<a href="/register">Register</a>
				`))
			}

			// Render each post with its author, categories, reactions, and comments.
			err = renderPosts(w, db, posts, currentUser)
			if err != nil {
				fmt.Println("renderPosts error:", err)
				http.Error(w, "failed to render posts", http.StatusInternalServerError)
				return
			}

			return
		}

		// POSTs and other methods should use the dedicated action endpoints.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
