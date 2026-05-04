package handlers

import (
	"database/sql"
	"net/http"
)

func HomeHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {

			cookie, err := r.Cookie("session")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			w.Write([]byte("you are logged in, user id: " + cookie.Value))

			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
