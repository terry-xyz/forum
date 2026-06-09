package handlers

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"forum/database"
	"forum/helpers"
	"html"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const (
	sessionCookieMaxAge = 60 * 60 * 24
	preAuthCSRFCookie   = "pre_auth_csrf"
)

func secureCookieForRequest(r *http.Request) bool {
	if r.TLS != nil {
		return true
	}

	for _, proto := range strings.Split(r.Header.Get("X-Forwarded-Proto"), ",") {
		if strings.EqualFold(strings.TrimSpace(proto), "https") {
			return true
		}
	}

	for _, forwarded := range r.Header.Values("Forwarded") {
		for _, part := range strings.Split(forwarded, ";") {
			keyValue := strings.SplitN(strings.TrimSpace(part), "=", 2)
			if len(keyValue) == 2 &&
				strings.EqualFold(keyValue[0], "proto") &&
				strings.EqualFold(strings.Trim(keyValue[1], `"`), "https") {
				return true
			}
		}
	}

	return false
}

func issuePreAuthCSRFToken(w http.ResponseWriter, r *http.Request) (string, error) {
	token, err := helpers.GenerateSessionID()
	if err != nil {
		return "", err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     preAuthCSRFCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   secureCookieForRequest(r),
		SameSite: http.SameSiteLaxMode,
	})

	return token, nil
}

func validPreAuthCSRFToken(r *http.Request) bool {
	cookie, err := r.Cookie(preAuthCSRFCookie)
	if err != nil || cookie.Value == "" {
		return false
	}

	submittedToken := r.PostForm.Get("csrf_token")
	if submittedToken == "" || len(submittedToken) != len(cookie.Value) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(submittedToken), []byte(cookie.Value)) == 1
}

// RegisterHandler serves the registration form and creates new user accounts.
func RegisterHandler(db *sql.DB) http.HandlerFunc {

	// Return a closure so the shared database dependency is captured once when
	// routes are registered.
	return func(w http.ResponseWriter, r *http.Request) {

		// GET renders the minimal registration form directly from the handler.
		if r.Method == http.MethodGet {
			csrfToken, err := issuePreAuthCSRFToken(w, r)
			if err != nil {
				http.Error(w, "unable to render form", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/register">
				<input type="hidden" name="csrf_token" value="` + html.EscapeString(csrfToken) + `">
				Email: <input type="email" name="email"><br>
				Username: <input name="username"><br>
				Password: <input type="password" name="password"><br>
				<button type="submit">Register</button>
			</form>`))
			return
		}

		// POST consumes submitted account details and attempts to persist them.
		if r.Method == http.MethodPost {
			if !parseLimitedForm(w, r) {
				return
			}
			if !validPreAuthCSRFToken(r) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			// FormValue parses form data on demand and returns an empty string
			// when a field is absent.
			email := strings.TrimSpace(r.FormValue("email"))
			username := strings.TrimSpace(r.FormValue("username"))
			password := strings.TrimSpace(r.FormValue("password"))

			_, err := mail.ParseAddress(email)

			if err != nil {
				http.Error(w, "invalid email format", http.StatusBadRequest)
				return
			}
			if username == "" {
				http.Error(w, "username cannot be empty", http.StatusBadRequest)
				return
			}
			if exceedsCharacterLimit(username, maxUsernameChars) {
				http.Error(w, "username cannot exceed 280 characters", http.StatusBadRequest)
				return
			}
			if len(password) < 8 {
				http.Error(w, "password must be at least 8 characters long", http.StatusBadRequest)
				return
			}

			// Store only a bcrypt hash so a database leak does not expose raw passwords.
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "unable to create user", http.StatusInternalServerError)
				return
			}

			// The database layer owns uniqueness enforcement through constraints.
			err = database.CreateUser(db, email, username, string(hashedPassword))
			if err != nil {
				if isConstraintError(err) {
					// Use the same public response as a successful registration so
					// attackers cannot test whether an email or username exists.
					http.Redirect(w, r, "/register", http.StatusSeeOther)
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
			csrfToken, err := issuePreAuthCSRFToken(w, r)
			if err != nil {
				http.Error(w, "unable to render form", http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`
			<form method="POST" action="/login">
				<input type="hidden" name="csrf_token" value="` + html.EscapeString(csrfToken) + `">
				Email: <input type="email" name="email"><br>
				Password: <input type="password" name="password"><br>
				<button type="submit">Login</button>
			</form>`))
			return
		}

		// POST validates submitted credentials against the users table.
		if r.Method == http.MethodPost {
			if !parseLimitedForm(w, r) {
				return
			}
			if !validPreAuthCSRFToken(r) {
				http.Error(w, "invalid csrf token", http.StatusForbidden)
				return
			}

			// Read the two fields the form submits; missing fields become empty
			// strings and naturally fail the credential check.
			email := strings.TrimSpace(r.FormValue("email"))
			password := strings.TrimSpace(r.FormValue("password"))

			_, err := mail.ParseAddress(email)

			if err != nil {
				http.Error(w, "invalid email format", http.StatusBadRequest)
				return
			}
			if len(password) < 8 {
				http.Error(w, "password must be at least 8 characters long", http.StatusBadRequest)
				return
			}

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

			err = database.DeleteSessionsByUserID(db, user.ID)
			if err != nil {
				http.Error(w, "unable to create session", http.StatusInternalServerError)
				return
			}

			sessionID, err := helpers.GenerateSessionID()
			if err != nil {
				http.Error(w, "unable to create session", http.StatusInternalServerError)
				return
			}
			csrfToken, err := helpers.GenerateSessionID()
			if err != nil {
				http.Error(w, "unable to create session", http.StatusInternalServerError)
				return
			}

			expiresAt := time.Now().Add(time.Second * sessionCookieMaxAge)

			err = database.CreateSession(db, sessionID, user.ID, csrfToken, expiresAt)
			if err != nil {
				http.Error(w, "unable to create session", http.StatusInternalServerError)
				return
			}

			// Store only the random session ID in the browser; the database maps
			// it back to the user for later requests.
			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				MaxAge:   sessionCookieMaxAge,
				HttpOnly: true,
				Secure:   secureCookieForRequest(r),
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

// LogoutHandler clears the session cookie, invalidates its database row, and returns the user to the home page.
func LogoutHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			http.Error(w, "invalid csrf token", http.StatusForbidden)
			return
		}

		if err := database.DeleteSession(db, cookie.Value); err != nil {
			http.Error(w, "unable to log out", http.StatusInternalServerError)
			return
		}

		// Reissue the same cookie name with MaxAge -1 and an old Expires value
		// so browsers remove it from storage.
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Expires:  time.Unix(0, 0), // Some browsers are stricter, so we make it extra clear the cookie is expired.
			HttpOnly: true,
			Secure:   secureCookieForRequest(r),
			SameSite: http.SameSiteLaxMode,
		})

		// Redirect to home; if the cookie was cleared, HomeHandler will require
		// the user to authenticate again.
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
