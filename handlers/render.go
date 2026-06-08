package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"forum/models"
)

type renderedPost struct {
	ID              int
	Title           string
	Content         string
	AuthorName      string
	Categories      []models.Category
	Likes           int
	Dislikes        int
	IsAuthenticated bool
	CanDelete       bool
	Comments        []renderedComment
}

type renderedComment struct {
	ID              int
	AuthorName      string
	Content         string
	Likes           int
	Dislikes        int
	IsAuthenticated bool
	CanDelete       bool
}

// buildRenderedPosts resolves feed data into the view model used by the home template.
func buildRenderedPosts(db *sql.DB, posts []models.Post, currentUser *models.User) ([]renderedPost, error) {
	renderedPosts := make([]renderedPost, 0, len(posts))

	for _, p := range posts {
		// Resolve author IDs while rendering so the UI can show usernames.
		author, err := database.GetUserByID(db, p.AuthorID)
		if err != nil {
			return nil, err
		}
		if author == nil {
			return nil, errors.New("post author not found")
		}

		// Reaction totals are stored as separate rows and counted per type.
		postLikes, err := database.CountPostReactions(db, p.ID, "like")
		if err != nil {
			return nil, err
		}
		postDislikes, err := database.CountPostReactions(db, p.ID, "dislike")
		if err != nil {
			return nil, err
		}

		// Categories are rendered as inline labels under each post.
		categories, err := database.GetCategoriesByPostID(db, p.ID)
		if err != nil {
			return nil, err
		}

		// Load comments and their authors before writing comment markup so
		// any database error can stop the response consistently.
		comments, err := database.GetCommentsByPostID(db, p.ID)
		if err != nil {
			return nil, err
		}
		renderedComments := make([]renderedComment, 0, len(comments))
		for _, c := range comments {
			// Each comment stores only author_id, so look up the display name.
			commentAuthor, err := database.GetUserByID(db, c.AuthorID)
			if err != nil {
				return nil, err
			}
			if commentAuthor == nil {
				return nil, errors.New("comment author not found")
			}
			commentLikes, err := database.CountCommentReactions(db, c.ID, "like")
			if err != nil {
				return nil, err
			}
			commentDislikes, err := database.CountCommentReactions(db, c.ID, "dislike")
			if err != nil {
				return nil, err
			}

			renderedComments = append(renderedComments, renderedComment{
				ID:              c.ID,
				AuthorName:      commentAuthor.Username,
				Content:         c.Content,
				Likes:           commentLikes,
				Dislikes:        commentDislikes,
				IsAuthenticated: currentUser != nil,
				CanDelete:       currentUser != nil && currentUser.ID == c.AuthorID,
			})
		}

		renderedPosts = append(renderedPosts, renderedPost{
			ID:              p.ID,
			Title:           p.Title,
			Content:         p.Content,
			AuthorName:      author.Username,
			Categories:      categories,
			Likes:           postLikes,
			Dislikes:        postDislikes,
			IsAuthenticated: currentUser != nil,
			CanDelete:       currentUser != nil && currentUser.ID == p.AuthorID,
			Comments:        renderedComments,
		})
	}

	return renderedPosts, nil
}
