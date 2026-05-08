package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"forum/models"
	"net/http"
	"strconv"
	"strings"
)

// renderPosts writes feed markup for posts after loading related authors, categories, comments, and reactions.
func renderPosts(w http.ResponseWriter, db *sql.DB, posts []models.Post) error {
	for _, p := range posts {
		// Resolve author IDs while rendering so the UI can show usernames.
		author, err := database.GetUserByID(db, p.AuthorID)
		if err != nil {
			return err
		}
		if author == nil {
			return errors.New("post author not found")
		}

		// Reaction totals are stored as separate rows and counted per type.
		postLikes, err := database.CountPostReactions(db, p.ID, "like")
		if err != nil {
			return err
		}
		postDislikes, err := database.CountPostReactions(db, p.ID, "dislike")
		if err != nil {
			return err
		}

		// Categories are rendered as inline labels under each post.
		categories, err := database.GetCategoriesByPostID(db, p.ID)
		if err != nil {
			return err
		}
		var categoryHTML strings.Builder
		for _, c := range categories {
			categoryHTML.WriteString("<span>" + c.Name + "</span> ")
		}

		// Load comments and their authors before writing comment markup so
		// any database error can stop the response consistently.
		comments, err := database.GetCommentsByPostID(db, p.ID)
		if err != nil {
			return err
		}
		commentAuthors := make([]string, 0, len(comments))
		for _, c := range comments {
			// Each comment stores only author_id, so look up the display name.
			commentAuthor, err := database.GetUserByID(db, c.AuthorID)
			if err != nil {
				return err
			}
			if commentAuthor == nil {
				return errors.New("comment author not found")
			}
			commentAuthors = append(commentAuthors, commentAuthor.Username)
		}

		// Write the post body, reaction forms, and comment form in one block.
		w.Write([]byte(
			"<h3>" + p.Title + "</h3>" +
				"<p>" + p.Content + "</p>" +
				"<small>Author: " + author.Username + "</small>" +
				"<p>Categories: " + categoryHTML.String() + "</p>" +
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
				return err
			}
			commentDislikes, err := database.CountCommentReactions(db, c.ID, "dislike")
			if err != nil {
				return err
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

	return nil
}
