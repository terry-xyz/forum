package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"
	"net/http"
)

// main is the application entry point: it prepares shared dependencies,
// registers routes, and starts the HTTP server.
func main() {
	// Open the SQLite database and run the schema before accepting requests.
	db, err := database.InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	// Wire each URL path to a handler. Handlers that need persistence receive
	// the same database handle so requests share one configured connection pool.
	http.HandleFunc("/", handlers.HomeHandler(db))
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/logout", handlers.LogoutHandler())
	http.HandleFunc("/create-post", handlers.CreatePostHandler(db))
	http.HandleFunc("/comment", handlers.CreateCommentHandler(db))
	http.HandleFunc("/post-reaction", handlers.ReactPostHandler(db))
	http.HandleFunc("/comment-reaction", handlers.ReactCommentHandler(db))

	// Serve the forum on the local development port and report startup/runtime
	// errors to stdout instead of panicking.
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
