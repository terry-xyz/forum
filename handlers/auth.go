package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"net/http"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
)

const sessionCookieMaxAge = 60 * 60 * 24

// RegisterHandler serves the registration form and creates new user accounts.
func RegisterHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/register">
				Email: <input type="email" name="email"><br>
				Username: <input name="username"><br>
				Password: <input type="password" name="password"><br>
				<button type="submit">Register</button>
			</form>`))
			return
		}

		if r.Method == http.MethodPost {
			email := r.FormValue("email")
			username := r.FormValue("username")
			password := r.FormValue("password")

			err := database.CreateUser(db, email, username, password)
			if err != nil {
				// Duplicate email or username values are reported by SQLite as constraint errors.
				if isConstraintError(err) {
					http.Error(w, "email or username is already taken", http.StatusConflict)
					return
				}

				http.Error(w, "unable to create user", http.StatusInternalServerError)
				return
			}
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func isConstraintError(err error) bool {
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint
}

// LoginHandler serves the login form and checks submitted credentials.
func LoginHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/login">
				Email: <input type="email" name="email"><br>
				Password: <input type="password" name="password"><br>
				<button type="submit">Login</button>
			</form>`))
			return
		}

		if r.Method == http.MethodPost {
			email := r.FormValue("email")
			password := r.FormValue("password")

			user, err := database.GetUserByEmail(db, email)
			if err != nil {
				http.Error(w, "unable to log in", http.StatusInternalServerError)
				return
			}

			// Use one message for both cases so the response does not reveal which emails exist.
			if user == nil || user.Password != password {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}

			// sessionID, err := helpers.GenerateSessionID()

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    strconv.Itoa(user.ID),
				Path:     "/",
				MaxAge:   sessionCookieMaxAge,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func LogoutHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0), // Some browsers are stricter, so we make it extra clear the cookie is expired.
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
