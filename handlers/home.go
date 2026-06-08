package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"forum/templates"
	"net/http"
	"strconv"
)

type homeView struct {
	Categories   []models.Category
	CurrentUser  *models.User
	CSRFToken    string
	Posts        []renderedPost
	EmptyMessage string
}

func renderHomePage(w http.ResponseWriter, db *sql.DB, posts []models.Post, currentUser *models.User, emptyMessage string, csrfToken string) error {
	allCategories, err := database.GetAllCategories(db)
	if err != nil {
		return err
	}
	renderedPosts, err := buildRenderedPosts(db, posts, currentUser, csrfToken)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	return templates.Home.Execute(w, homeView{
		Categories:   allCategories,
		CurrentUser:  currentUser,
		CSRFToken:    csrfToken,
		Posts:        renderedPosts,
		EmptyMessage: emptyMessage,
	})
}

// HomeHandler renders the forum home page for guests and adds user actions when a session is valid.
func HomeHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle for all requests served by this route.
	return func(w http.ResponseWriter, r *http.Request) {

		// The home page is currently read-only, so only GET is supported.
		if r.Method == http.MethodGet {

			var currentUser *models.User
			csrfToken := ""

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
					csrfToken, err = database.GetCSRFTokenBySessionID(db, cookie.Value)
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

			// Render the full page from the file-backed template.
			if err := renderHomePage(w, db, posts, currentUser, "", csrfToken); err != nil {
				http.Error(w, "failed to render home page", http.StatusInternalServerError)
				return
			}

			return
		}

		// POSTs and other methods should use the dedicated action endpoints.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
