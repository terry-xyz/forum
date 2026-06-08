package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestCreatePostAndGetAllPosts covers inserting a post and reading it back.
func TestCreatePostAndGetAllPosts(t *testing.T) {
	// Start each database test from a clean schema.
	db := openTestDB(t)

	// Create an author because posts require an author_id value.
	if err := CreateUser(db, "author@example.com", "author", "password"); err != nil {
		t.Fatal(err)
	}

	// Look up the generated user ID for use as the post author.
	user, err := GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}

	// Insert one post through the public database helper under test.
	if _, err := CreatePost(db, user.ID, "First post", "Hello forum"); err != nil {
		t.Fatal(err)
	}

	// Read all posts back and assert the inserted row is represented correctly.
	posts, err := GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}

	// The fresh database should contain exactly the post inserted above.
	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}

	// Check each meaningful field so scan order and inserted values are covered.
	post := posts[0]
	if post.AuthorID != user.ID {
		t.Fatalf("author id = %d, want %d", post.AuthorID, user.ID)
	}
	if post.Title != "First post" {
		t.Fatalf("title = %q, want First post", post.Title)
	}
	if post.Content != "Hello forum" {
		t.Fatalf("content = %q, want Hello forum", post.Content)
	}
	if post.CreatedAt.IsZero() {
		t.Fatal("created at should be set")
	}
}

// TestGetUserByIDReturnsNilForMissingUser documents the nil-on-miss contract.
func TestGetUserByIDReturnsNilForMissingUser(t *testing.T) {
	// No users are inserted, so any lookup should miss.
	db := openTestDB(t)

	// Missing rows should not be returned as errors to callers.
	user, err := GetUserByID(db, 999)
	if err != nil {
		t.Fatal(err)
	}
	if user != nil {
		t.Fatalf("user = %#v, want nil", user)
	}
}

// openTestDB creates a temporary SQLite database initialized with schema.sql.
func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Use t.TempDir so each test has an isolated database file.
	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db?_foreign_keys=on")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	// Load the same schema file used by the database package.
	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}

	// Apply the schema before returning the handle to the test.
	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	return db
}
