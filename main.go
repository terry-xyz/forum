package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	defaultPort         = "8080"
	defaultDatabasePath = "forum.db"
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

func newAppMux() *http.ServeMux {
	mux := http.NewServeMux()
	registerStaticHandlers(mux)
	return mux
}

func registerStaticHandlers(mux *http.ServeMux) {
	fileServer := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
}

func newHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           securityHeaders(handler),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
}

func configuredServerAddress() string {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = defaultPort
	}
	if strings.HasPrefix(port, ":") || strings.Contains(port, ":") {
		return port
	}
	return ":" + port
}

func configuredDatabasePath() string {
	databasePath := strings.TrimSpace(os.Getenv("DATABASE_PATH"))
	if databasePath == "" {
		return defaultDatabasePath
	}
	return databasePath
}

// main prepares the database-backed handlers and starts the HTTP server.
func main() {
	// InitDB opens SQLite and applies the schema before handlers can use it.
	db, err := database.InitDB(configuredDatabasePath())
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
	mux := newAppMux()
	mux.HandleFunc("/", handlers.HomeHandler(db))
	mux.HandleFunc("/register", handlers.RegisterHandler(db))
	mux.HandleFunc("/login", handlers.LoginHandler(db))
	mux.HandleFunc("/logout", handlers.LogoutHandler(db))
	mux.HandleFunc("/create-post", handlers.CreatePostHandler(db))
	mux.HandleFunc("/comment", handlers.CreateCommentHandler(db))
	mux.HandleFunc("/post-reaction", handlers.ReactPostHandler(db))
	mux.HandleFunc("/comment-reaction", handlers.ReactCommentHandler(db))
	mux.HandleFunc("/my-posts", handlers.MyPostsHandler(db))
	mux.HandleFunc("/liked-posts", handlers.LikedPostsHandler(db))
	mux.HandleFunc("/delete-comment", handlers.DeleteCommentHandler(db))
	mux.HandleFunc("/delete-post", handlers.DeletePostHandler(db))

	addr := configuredServerAddress()
	// Print before ListenAndServe because it blocks until the server stops or fails.
	fmt.Printf("Starting server on %s\n", serverURL(addr))

	server := newHTTPServer(addr, mux)
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}

func serverURL(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "http://localhost" + addr
	}
	return "http://" + addr
}
