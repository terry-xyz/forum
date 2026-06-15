package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"forum/models"
	"forum/templates"
	"net/http"
	"net/url"
	"strconv"
)

const (
	feedPageSize            = 20
	feedCommentPreviewLimit = 5
)

type homeView struct {
	Categories         []models.Category
	CurrentUser        *models.User
	CSRFToken          string
	Posts              []renderedPost
	EmptyMessage       string
	SelectedCategoryID int
	HasPrev            bool
	HasNext            bool
	PrevURL            string
	NextURL            string
}

func renderHomePage(w http.ResponseWriter, db *sql.DB, posts []models.Post, currentUser *models.User, emptyMessage string, csrfToken string, selectedCategoryID int, pagination paginationView, commentErrorPostID int, commentError string) error {
	allCategories, err := database.GetAllCategories(db)
	if err != nil {
		return err
	}
	renderedPosts, err := buildRenderedPosts(db, posts, currentUser, csrfToken, commentErrorPostID, commentError)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "text/html")
	return templates.Home.Execute(w, homeView{
		Categories:         allCategories,
		CurrentUser:        currentUser,
		CSRFToken:          csrfToken,
		Posts:              renderedPosts,
		EmptyMessage:       emptyMessage,
		SelectedCategoryID: selectedCategoryID,
		HasPrev:            pagination.HasPrev,
		HasNext:            pagination.HasNext,
		PrevURL:            pagination.PrevURL,
		NextURL:            pagination.NextURL,
	})
}

type paginationView struct {
	HasPrev bool
	HasNext bool
	PrevURL string
	NextURL string
}

func parsePage(r *http.Request) (int, error) {
	pageValue := r.URL.Query().Get("page")
	if pageValue == "" {
		return 1, nil
	}

	page, err := strconv.Atoi(pageValue)
	if err != nil {
		return 0, err
	}
	if page < 1 {
		return 0, errors.New("page must be positive")
	}

	return page, nil
}

func paginationForRequest(r *http.Request, page int, hasNext bool) paginationView {
	return paginationView{
		HasPrev: page > 1,
		HasNext: hasNext,
		PrevURL: pageURL(r.URL.Path, r.URL.Query(), page-1),
		NextURL: pageURL(r.URL.Path, r.URL.Query(), page+1),
	}
}

func pageURL(path string, values url.Values, page int) string {
	if page < 1 {
		page = 1
	}

	copiedValues := make(url.Values, len(values))
	for key, value := range values {
		copiedValues[key] = append([]string(nil), value...)
	}
	if page == 1 {
		copiedValues.Del("page")
	} else {
		copiedValues.Set("page", strconv.Itoa(page))
	}

	query := copiedValues.Encode()
	if query == "" {
		return path
	}
	return path + "?" + query
}

func trimPage(posts []models.Post) ([]models.Post, bool) {
	if len(posts) <= feedPageSize {
		return posts, false
	}

	return posts[:feedPageSize], true
}

// HomeHandler renders the forum home page for guests and adds user actions when a session is valid.
func HomeHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle for all requests served by this route.
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			renderErrorPage(w, http.StatusNotFound, "Page not found", "The page you requested does not exist.")
			return
		}

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
					renderHTTPError(w, http.StatusInternalServerError, "internal error")
					return
				}
				if userID != 0 {
					currentUser, err = database.GetUserByID(db, userID)
					if err != nil {
						renderHTTPError(w, http.StatusInternalServerError, "internal error")
						return
					}
					csrfToken, err = database.GetCSRFTokenBySessionID(db, cookie.Value)
					if err != nil {
						renderHTTPError(w, http.StatusInternalServerError, "internal error")
						return
					}
				}
			}

			page, err := parsePage(r)
			if err != nil {
				renderHTTPError(w, http.StatusBadRequest, "invalid page")
				return
			}
			offset := (page - 1) * feedPageSize

			// Choose between the default feed and a category-filtered feed based
			// on the optional query string.
			var posts []models.Post
			selectedCategoryID := 0
			categoryIDStr := r.URL.Query().Get("category_id")
			if categoryIDStr == "" {
				// No filter means the newest feed page should be shown.
				posts, err = database.GetAllPostsPage(db, feedPageSize+1, offset)
				if err != nil {
					renderHTTPError(w, http.StatusInternalServerError, "internal server error")
					return
				}
			} else {
				// Bad query values are client errors because category_id is part of
				// the request URL.
				categoryID, err := strconv.Atoi(categoryIDStr)
				if err != nil {
					renderHTTPError(w, http.StatusBadRequest, "invalid category id")
					return
				}
				selectedCategoryID = categoryID
				// A valid category ID narrows the feed to posts linked to it.
				posts, err = database.GetPostsByCategoryIDPage(db, categoryID, feedPageSize+1, offset)
				if err != nil {
					renderHTTPError(w, http.StatusInternalServerError, "internal server error")
					return
				}
			}
			posts, hasNext := trimPage(posts)

			// Render the full page from the file-backed template.
			if err := renderHomePage(w, db, posts, currentUser, "", csrfToken, selectedCategoryID, paginationForRequest(r, page, hasNext), 0, ""); err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "failed to render home page")
				return
			}

			return
		}

		// POSTs and other methods should use the dedicated action endpoints.
		renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
