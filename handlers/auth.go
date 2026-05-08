package handlers

import (
	"database/sql"
	"errors"
	"forum/database"
	"net/http"
	"strconv"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const sessionCookieMaxAge = 60 * 60 * 24

// RegisterHandler serves the registration form and creates new user accounts.
func RegisterHandler(db *sql.DB) http.HandlerFunc {

	// Return a closure so the shared database dependency is captured once when
	// routes are registered.
	return func(w http.ResponseWriter, r *http.Request) {

		// GET renders the minimal registration form directly from the handler.
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

		// POST consumes submitted account details and attempts to persist them.
		if r.Method == http.MethodPost {
			// FormValue parses form data on demand and returns an empty string
			// when a field is absent.
			email := r.FormValue("email")
			username := r.FormValue("username")
			password := r.FormValue("password")

			// Store only a bcrypt hash so a database leak does not expose raw passwords.
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "unable to create user", http.StatusInternalServerError)
				return
			}

			// The database layer owns uniqueness enforcement through constraints.
			err = database.CreateUser(db, email, username, string(hashedPassword))
			if err != nil {
				// Duplicate email or username values are reported by SQLite as constraint errors.
				if isConstraintError(err) {
					http.Error(w, "email or username is already taken", http.StatusConflict)
					return
				}

				http.Error(w, "unable to create user", http.StatusInternalServerError)
				return
			}

			// Redirect after a successful POST so browser refresh does not submit
			// the same registration again.
			http.Redirect(w, r, "/register", http.StatusSeeOther)
			return
		}

		// Any method outside the two supported branches is rejected explicitly.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// isConstraintError reports whether an error came from a SQLite constraint.
func isConstraintError(err error) bool {
	// errors.As handles wrapped sqlite errors without depending on exact error
	// message strings.
	var sqliteErr sqlite3.Error
	return errors.As(err, &sqliteErr) && sqliteErr.Code == sqlite3.ErrConstraint
}

// LoginHandler serves the login form and checks submitted credentials.
func LoginHandler(db *sql.DB) http.HandlerFunc {

	// Capture db once and return a standard net/http handler function.
	return func(w http.ResponseWriter, r *http.Request) {

		// GET presents the login form without touching the database.
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

		// POST validates submitted credentials against the users table.
		if r.Method == http.MethodPost {
			// Read the two fields the form submits; missing fields become empty
			// strings and naturally fail the credential check.
			email := r.FormValue("email")
			password := r.FormValue("password")

			// Lookup by email first because email is the login identifier.
			user, err := database.GetUserByEmail(db, email)
			if err != nil {
				http.Error(w, "unable to log in", http.StatusInternalServerError)
				return
			}

			// Use the same response for unknown emails and bad passwords so login
			// attempts cannot reveal which accounts exist.
			if user == nil {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}

			// CompareHashAndPassword validates the submitted password against the
			// stored bcrypt hash without needing the original password.
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				http.Error(w, "invalid email or password", http.StatusUnauthorized)
				return
			}

			// Store the current user ID in an HttpOnly cookie. This is the
			// lightweight session mechanism used throughout the handlers.
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    strconv.Itoa(user.ID),
				Path:     "/",
				MaxAge:   sessionCookieMaxAge,
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			})
			// Send the user to the forum after the cookie has been set.
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Login only supports rendering and submitting the form.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// LogoutHandler clears the session cookie and returns the user to the home page.
func LogoutHandler() http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		// Reissue the same cookie name with MaxAge -1 and an old Expires value
		// so browsers remove it from storage.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0), // Some browsers are stricter, so we make it extra clear the cookie is expired.
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to home; if the cookie was cleared, HomeHandler will require
		// the user to authenticate again.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
