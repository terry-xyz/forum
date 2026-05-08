package main

import (
	"fmt"
	"forum/database"
	"os"
)

// main opens the forum database and fills it with deterministic demo data.
func main() {
	// InitDB applies the schema first so the seed can run on a fresh checkout.
	db, err := database.InitDB()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()

	// SeedFakeData is idempotent, so running this command repeatedly refreshes missing data only.
	if err := database.SeedFakeData(db); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println("seeded forum.db with fake forum data")
}
