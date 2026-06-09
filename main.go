package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"
	"net/http"
	"time"
)

const contentSecurityPolicy = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; object-src 'none'; upgrade-insecure-requests"

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", contentSecurityPolicy)
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

func newHTTPServer(addr string) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           securityHeaders(http.DefaultServeMux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

// main prepares the database-backed handlers and starts the HTTP server.
func main() {
	// InitDB opens SQLite and applies the schema before handlers can use it.
	db, err := database.InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	defer db.Close()

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

	server := newHTTPServer(defaultPort)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
