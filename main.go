package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"
	"net/http"
)

// main prepares the database-backed handlers and starts the HTTP server.
func main() {
	// InitDB opens SQLite and applies the schema before handlers can use it.
	db, err := database.InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	err = database.DeleteExpiredSessions(db)
	if err != nil {
		fmt.Println(err)
		return
	}

	// Each handler receives the shared database handle so all requests use the
	// same initialized connection.
	http.HandleFunc("/", handlers.HomeHandler(db))
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/logout", handlers.LogoutHandler(db))
	http.HandleFunc("/create-post", handlers.CreatePostHandler(db))
	http.HandleFunc("/comment", handlers.CreateCommentHandler(db))
	http.HandleFunc("/post-reaction", handlers.ReactPostHandler(db))
	http.HandleFunc("/comment-reaction", handlers.ReactCommentHandler(db))
	http.HandleFunc("/my-posts", handlers.MyPostsHandler(db))
	http.HandleFunc("/liked-posts", handlers.LikedPostsHandler(db))
	http.HandleFunc("/delete-comment", handlers.DeleteCommentHandler(db))
	http.HandleFunc("/delete-post", handlers.DeletePostHandler(db))

	defaultPort := ":8080"
	// Print before ListenAndServe because it blocks until the server stops or fails.
	fmt.Printf("Starting server on http://localhost%s\n", defaultPort)

	err = http.ListenAndServe(defaultPort, nil)
	if err != nil {
		fmt.Println(err)
	}
}
