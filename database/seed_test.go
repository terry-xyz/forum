package database

import (
	"database/sql"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestSeedFakeDataPopulatesForumTables verifies the seed creates every related forum table row.
func TestSeedFakeDataPopulatesForumTables(t *testing.T) {
	// Use a fresh schema so row counts reflect only this seed run.
	db := openTestDB(t)

	if err := SeedFakeData(db); err != nil {
		t.Fatal(err)
	}

	// Each table count proves the seed wrote users, posts, joins, comments, and reactions.
	expectedCounts := map[string]int{
		"users":             3,
		"posts":             4,
		"comments":          6,
		"post_categories":   6,
		"post_reactions":    6,
		"comment_reactions": 6,
	}

	for table, want := range expectedCounts {
		if got := countRows(t, db, table); got != want {
			t.Fatalf("%s count = %d, want %d", table, got, want)
		}
	}

	// Verify seeded passwords are usable by the same bcrypt login path as real registrations.
	user, err := GetUserByEmail(db, "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user == nil {
		t.Fatal("seeded alice user was not found")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte("password123")); err != nil {
		t.Fatalf("seeded password is not bcrypt-compatible: %v", err)
	}
}

// TestSeedFakeDataCanRunMoreThanOnce verifies seed inserts are idempotent.
func TestSeedFakeDataCanRunMoreThanOnce(t *testing.T) {
	// Start from an empty test database so the second run is the only duplicate path.
	db := openTestDB(t)

	if err := SeedFakeData(db); err != nil {
		t.Fatal(err)
	}
	if err := SeedFakeData(db); err != nil {
		t.Fatal(err)
	}

	// Counts should stay the same after the second run because helpers reuse existing rows.
	expectedCounts := map[string]int{
		"users":             3,
		"posts":             4,
		"comments":          6,
		"post_categories":   6,
		"post_reactions":    6,
		"comment_reactions": 6,
	}

	for table, want := range expectedCounts {
		if got := countRows(t, db, table); got != want {
			t.Fatalf("%s count after second seed = %d, want %d", table, got, want)
		}
	}
}

// countRows returns the number of rows in one trusted test table.
func countRows(t *testing.T, db *sql.DB, table string) int {
	t.Helper()

	var count int
	// Table names come from test constants above, so Sprintf does not include user input.
	if err := db.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", table)).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
