package main

import (
	"fmt"
	"forum/database"
	"forum/handlers"
	"net/http"
)

func main() {
	db, err := database.InitDB()
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/", handlers.HomeHandler(db))
	http.HandleFunc("/register", handlers.RegisterHandler(db))
	http.HandleFunc("/login", handlers.LoginHandler(db))
	http.HandleFunc("/logout", handlers.LogoutHandler())
	http.HandleFunc("/create-post", handlers.CreatePostHandler(db))

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println(err)
	}
}
