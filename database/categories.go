package database

import (
	"database/sql"
	"forum/models"
)

// GetAllCategories returns every category in display order.
func GetAllCategories(db *sql.DB) ([]models.Category, error) {

	// Alphabetical order gives forms and filters a stable, predictable layout.
	query := "SELECT id, name FROM categories ORDER BY name ASC"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	// Ensure the database cursor is closed after the category list is read.
	defer rows.Close()

	var categories []models.Category

	// Convert each row into the small category model used by handlers.
	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	// Check for delayed scan/iteration failures.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// GetCategoriesByPostID returns the labels attached to one post.
func GetCategoriesByPostID(db *sql.DB, postID int) ([]models.Category, error) {

	// Resolve the many-to-many link table back into category rows for display.
	query := `
		SELECT c.id, c.name
		FROM categories c
		JOIN post_categories pc ON pc.category_id = c.id
		WHERE pc.post_id = ?
		ORDER BY c.name ASC

	`
	rows, err := db.Query(query, postID)
	if err != nil {
		return nil, err
	}

	// Close the joined result set once all category names are scanned.
	defer rows.Close()

	var categories []models.Category

	// Preserve only category fields; the link table columns are not needed by UI.
	for rows.Next() {
		var category models.Category

		err := rows.Scan(&category.ID, &category.Name)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	// Return any error raised during row iteration.
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

// AddCategoriesToPost creates post/category links for a newly created post.
func AddCategoriesToPost(db *sql.DB, postID int, categoryIDs []int) error {

	// Insert each selected category into the join table. The UNIQUE constraint
	// prevents duplicate links for the same post/category pair.
	query := "INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)"

	for _, categoryID := range categoryIDs {
		_, err := db.Exec(query, postID, categoryID)
		if err != nil {
			return err
		}
	}

	// Success means every requested relationship was written.
	return nil
}
