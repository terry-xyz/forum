package handlers

import (
	"database/sql"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
)

// HomeHandler renders the forum home page for users with a valid session cookie.
func HomeHandler(db *sql.DB) http.HandlerFunc {

	// Capture the database handle for all requests served by this route.
	return func(w http.ResponseWriter, r *http.Request) {

		// The home page is currently read-only, so only GET is supported.
		if r.Method == http.MethodGet {

			// Require a session cookie before loading posts or rendering the page.
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

			// Confirm that the user referenced by the session still exists.
			user, err := database.GetUserByID(db, cookieID)
			if err != nil {
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
			if user == nil {
				http.Error(w, "invalid session", http.StatusUnauthorized)
				return
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

			// Render each post with its author, categories, reactions, and comments.
			for _, p := range posts {
				// Resolve author IDs while rendering so the UI can show usernames.
				author, err := database.GetUserByID(db, p.AuthorID)
				if err != nil {
					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				if author == nil {
					http.Error(w, "invalid post author", http.StatusInternalServerError)
					return
				}

				// Reaction totals are stored as separate rows and counted per type.
				postLikes, err := database.CountPostReactions(db, p.ID, "like")
				if err != nil {
					http.Error(w, "internal server error, couldnt get likes amount", http.StatusInternalServerError)
					return
				}
				postDislikes, err := database.CountPostReactions(db, p.ID, "dislike")
				if err != nil {
					http.Error(w, "internal server error, couldnt get dislikes amount", http.StatusInternalServerError)
					return
				}

				// Categories are rendered as inline labels under each post.
				categories, err := database.GetCategoriesByPostID(db, p.ID)
				if err != nil {
					http.Error(w, "failed to load categories", http.StatusInternalServerError)
					return
				}
				categoryHTML := ""
				for _, c := range categories {
					categoryHTML += "<span>" + c.Name + "</span> "
				}

				// Load comments and their authors before writing comment markup so
				// any database error can stop the response consistently.
				comments, err := database.GetCommentsByPostID(db, p.ID)
				if err != nil {
					http.Error(w, "failed to load comments", http.StatusInternalServerError)
					return
				}
				commentAuthors := make([]string, 0, len(comments))
				for _, c := range comments {
					// Each comment stores only author_id, so look up the display name.
					commentAuthor, err := database.GetUserByID(db, c.AuthorID)
					if err != nil {
						http.Error(w, "internal server error", http.StatusInternalServerError)
						return
					}
					if commentAuthor == nil {
						http.Error(w, "invalid comment author", http.StatusInternalServerError)
						return
					}
					commentAuthors = append(commentAuthors, commentAuthor.Username)
				}

				// Write the post body, reaction forms, and comment form in one block.
				w.Write([]byte(
					"<h3>" + p.Title + "</h3>" +
						"<p>" + p.Content + "</p>" +
						"<small>Author: " + author.Username + "</small>" +
						"<p>Categories: " + categoryHTML + "</p>" +
						"<p>Likes: " + strconv.Itoa(postLikes) + " | Dislikes: " + strconv.Itoa(postDislikes) + "</p>" +
						`<form method="POST" action="/post-reaction">
							<input type="hidden" name="post_id" value="` + strconv.Itoa(p.ID) + `">
							<input type="hidden" name="reaction_type" value="like">
							<button type="submit">Like</button>
						</form>

						<form method="POST" action="/post-reaction">
							<input type="hidden" name="post_id" value="` + strconv.Itoa(p.ID) + `">
							<input type="hidden" name="reaction_type" value="dislike">
							<button type="submit">Dislike</button>
						</form>` +
						`<form method="POST" action="/comment">
							<input type="hidden" name="post_id" value="` + strconv.Itoa(p.ID) + `">
							<textarea name="content"></textarea>
							<button type="submit">Comment</button>
						</form>`,
				))
				// Render comments below their owning post, preserving the same index
				// into commentAuthors that was built during validation.
				for i, c := range comments {
					// Counts are fetched per comment so each comment has independent
					// like/dislike totals.
					commentLikes, err := database.CountCommentReactions(db, c.ID, "like")
					if err != nil {
						http.Error(w, "internal server error, couldnt get likes amount", http.StatusInternalServerError)
						return
					}
					commentDislikes, err := database.CountCommentReactions(db, c.ID, "dislike")
					if err != nil {
						http.Error(w, "internal server error, couldnt get dislikes amount", http.StatusInternalServerError)
						return
					}
					w.Write([]byte(
						"<h5>" + commentAuthors[i] + "</h5>" +
							"<p>" + c.Content + "</p>" +
							"<p>Likes: " + strconv.Itoa(commentLikes) + " | Dislikes: " + strconv.Itoa(commentDislikes) + "</p>" +
							`<form method="POST" action="/comment-reaction">
								<input type="hidden" name="comment_id" value="` + strconv.Itoa(c.ID) + `">
								<input type="hidden" name="reaction_type" value="like">
								<button type="submit">Like</button>
							</form>

							<form method="POST" action="/comment-reaction">
								<input type="hidden" name="comment_id" value="` + strconv.Itoa(c.ID) + `">
								<input type="hidden" name="reaction_type" value="dislike">
								<button type="submit">Dislike</button>
							</form>`,
					))
				}
				// Separate posts visually in the simple HTML output.
				w.Write([]byte("<hr>"))
			}

			return
		}

		// POSTs and other methods should use the dedicated action endpoints.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
