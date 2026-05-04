package database

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCreatePostAndGetAllPosts(t *testing.T) {
	db := openTestDB(t)

	if err := CreateUser(db, "author@example.com", "author", "password"); err != nil {
		t.Fatal(err)
	}

	user, err := GetUserByEmail(db, "author@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("user was not created")
	}

	if err := CreatePost(db, user.ID, "First post", "Hello forum"); err != nil {
		t.Fatal(err)
	}

	posts, err := GetAllPosts(db)
	if err != nil {
		t.Fatal(err)
	}

	if len(posts) != 1 {
		t.Fatalf("posts = %d, want 1", len(posts))
	}

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

func TestGetUserByIDReturnsNilForMissingUser(t *testing.T) {
	db := openTestDB(t)

	user, err := GetUserByID(db, 999)
	if err != nil {
		t.Fatal(err)
	}
	if user != nil {
		t.Fatalf("user = %#v, want nil", user)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	db, err := sql.Open("sqlite3", t.TempDir()+"/forum.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	schema, err := os.ReadFile("schema.sql")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatal(err)
	}

	return db
}
