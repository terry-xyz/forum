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
	CSRFToken       string
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
	CSRFToken       string
}

// buildRenderedPosts resolves feed data into the view model used by the home template.
func buildRenderedPosts(db *sql.DB, posts []models.Post, currentUser *models.User, csrfToken string) ([]renderedPost, error) {
	renderedPosts := make([]renderedPost, 0, len(posts))
	if len(posts) == 0 {
		return renderedPosts, nil
	}

	postIDs := make([]int, 0, len(posts))
	authorIDs := make([]int, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
		authorIDs = append(authorIDs, post.AuthorID)
	}

	postAuthors, err := database.GetUsersByIDs(db, uniqueInts(authorIDs))
	if err != nil {
		return nil, err
	}
	postReactions, err := database.CountPostReactionsByPostIDs(db, postIDs)
	if err != nil {
		return nil, err
	}
	categoriesByPost, err := database.GetCategoriesByPostIDs(db, postIDs)
	if err != nil {
		return nil, err
	}
	commentsByPost, err := database.GetRecentCommentsByPostIDs(db, postIDs, feedCommentPreviewLimit)
	if err != nil {
		return nil, err
	}

	commentAuthorIDs := make([]int, 0)
	commentIDs := make([]int, 0)
	for _, comments := range commentsByPost {
		for _, comment := range comments {
			commentAuthorIDs = append(commentAuthorIDs, comment.AuthorID)
			commentIDs = append(commentIDs, comment.ID)
		}
	}
	commentAuthors, err := database.GetUsersByIDs(db, uniqueInts(commentAuthorIDs))
	if err != nil {
		return nil, err
	}
	commentReactions, err := database.CountCommentReactionsByCommentIDs(db, commentIDs)
	if err != nil {
		return nil, err
	}

	for _, p := range posts {
		author, ok := postAuthors[p.AuthorID]
		if !ok {
			return nil, errors.New("post author not found")
		}

		postReactionCounts := postReactions[p.ID]
		comments := commentsByPost[p.ID]
		renderedComments := make([]renderedComment, 0, len(comments))
		for _, c := range comments {
			commentAuthor, ok := commentAuthors[c.AuthorID]
			if !ok {
				return nil, errors.New("comment author not found")
			}
			commentReactionCounts := commentReactions[c.ID]

			renderedComments = append(renderedComments, renderedComment{
				ID:              c.ID,
				AuthorName:      commentAuthor.Username,
				Content:         c.Content,
				Likes:           commentReactionCounts.Likes,
				Dislikes:        commentReactionCounts.Dislikes,
				IsAuthenticated: currentUser != nil,
				CanDelete:       currentUser != nil && currentUser.ID == c.AuthorID,
				CSRFToken:       csrfToken,
			})
		}

		renderedPosts = append(renderedPosts, renderedPost{
			ID:              p.ID,
			Title:           p.Title,
			Content:         p.Content,
			AuthorName:      author.Username,
			Categories:      categoriesByPost[p.ID],
			Likes:           postReactionCounts.Likes,
			Dislikes:        postReactionCounts.Dislikes,
			IsAuthenticated: currentUser != nil,
			CanDelete:       currentUser != nil && currentUser.ID == p.AuthorID,
			CSRFToken:       csrfToken,
			Comments:        renderedComments,
		})
	}

	return renderedPosts, nil
}

func uniqueInts(values []int) []int {
	seen := make(map[int]bool, len(values))
	unique := make([]int, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	return unique
}
