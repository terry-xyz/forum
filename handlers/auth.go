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

func registerFormHTML(csrfToken string, email string, username string) string {
	return `
			<form method="POST" action="/register">
				<input type="hidden" name="csrf_token" value="` + html.EscapeString(csrfToken) + `">
				<label>Email <input type="email" name="email" value="` + html.EscapeString(email) + `"></label>
				<label>Username <input name="username" value="` + html.EscapeString(username) + `"></label>
				<label>Password <input type="password" name="password"></label>
				<button type="submit">Register</button>
			</form>`
}

func loginFormHTML(csrfToken string, email string) string {
	return `
			<form method="POST" action="/login">
				<input type="hidden" name="csrf_token" value="` + html.EscapeString(csrfToken) + `">
				<label>Email <input type="email" name="email" value="` + html.EscapeString(email) + `"></label>
				<label>Password <input type="password" name="password"></label>
				<button type="submit">Login</button>
			</form>`
}

func validEmailAddress(email string) bool {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return false
	}

	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" || strings.ContainsAny(email, " \t\r\n") {
		return false
	}

	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return false
	}

	for _, label := range strings.Split(domain, ".") {
		if label == "" {
			return false
		}
	}

	return true
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
				renderHTTPError(w, http.StatusInternalServerError, "unable to render form")
				return
			}

			renderFormPage(w, "Register", "", registerFormHTML(csrfToken, "", ""))
			return
		}

		// POST consumes submitted account details and attempts to persist them.
		if r.Method == http.MethodPost {
			if !parseLimitedForm(w, r) {
				return
			}
			if !validPreAuthCSRFToken(r) {
				renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
				return
			}

			// FormValue parses form data on demand and returns an empty string
			// when a field is absent.
			email := strings.TrimSpace(r.FormValue("email"))
			username := strings.TrimSpace(r.FormValue("username"))
			password := strings.TrimSpace(r.FormValue("password"))

			if !validEmailAddress(email) {
				renderFormPage(w, "Register", "invalid email format", registerFormHTML(r.PostForm.Get("csrf_token"), email, username))
				return
			}
			if username == "" {
				renderFormPage(w, "Register", "username cannot be empty", registerFormHTML(r.PostForm.Get("csrf_token"), email, username))
				return
			}
			if exceedsCharacterLimit(username, maxUsernameChars) {
				renderFormPage(w, "Register", "username cannot exceed 280 characters", registerFormHTML(r.PostForm.Get("csrf_token"), email, username))
				return
			}
			if len(password) < 8 {
				renderFormPage(w, "Register", "password must be at least 8 characters long", registerFormHTML(r.PostForm.Get("csrf_token"), email, username))
				return
			}

			// Store only a bcrypt hash so a database leak does not expose raw passwords.
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to create user")
				return
			}

			// The database layer owns uniqueness enforcement through constraints.
			err = database.CreateUser(db, email, username, string(hashedPassword))
			if err != nil {
				if isConstraintError(err) {
					// Use the same public response as a successful registration so
					// attackers cannot test whether an email or username exists.
					http.Redirect(w, r, "/login", http.StatusSeeOther)
					return
				}

				renderHTTPError(w, http.StatusInternalServerError, "unable to create user")
				return
			}

			// Redirect after a successful POST so browser refresh does not submit
			// the same registration again.
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Any method outside the two supported branches is rejected explicitly.
		renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
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
				renderHTTPError(w, http.StatusInternalServerError, "unable to render form")
				return
			}

			renderFormPage(w, "Login", "", loginFormHTML(csrfToken, ""))
			return
		}

		// POST validates submitted credentials against the users table.
		if r.Method == http.MethodPost {
			if !parseLimitedForm(w, r) {
				return
			}
			if !validPreAuthCSRFToken(r) {
				renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
				return
			}

			// Read the two fields the form submits; missing fields become empty
			// strings and naturally fail the credential check.
			email := strings.TrimSpace(r.FormValue("email"))
			password := strings.TrimSpace(r.FormValue("password"))

			_, err := mail.ParseAddress(email)

			if err != nil {
				renderFormPage(w, "Login", "invalid email format", loginFormHTML(r.PostForm.Get("csrf_token"), email))
				return
			}
			if len(password) < 8 {
				renderFormPage(w, "Login", "password must be at least 8 characters long", loginFormHTML(r.PostForm.Get("csrf_token"), email))
				return
			}

			// Lookup by email first because email is the login identifier.
			user, err := database.GetUserByEmail(db, email)
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to log in")
				return
			}

			// Use the same response for unknown emails and bad passwords so login
			// attempts cannot reveal which accounts exist.
			if user == nil {
				renderFormPage(w, "Login", "invalid email or password", loginFormHTML(r.PostForm.Get("csrf_token"), email))
				return
			}

			// CompareHashAndPassword validates the submitted password against the
			// stored bcrypt hash without needing the original password.
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				renderFormPage(w, "Login", "invalid email or password", loginFormHTML(r.PostForm.Get("csrf_token"), email))
				return
			}

			err = database.DeleteSessionsByUserID(db, user.ID)
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to create session")
				return
			}

			sessionID, err := helpers.GenerateSessionID()
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to create session")
				return
			}
			csrfToken, err := helpers.GenerateSessionID()
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to create session")
				return
			}

			expiresAt := time.Now().Add(time.Second * sessionCookieMaxAge)

			err = database.CreateSession(db, sessionID, user.ID, csrfToken, expiresAt)
			if err != nil {
				renderHTTPError(w, http.StatusInternalServerError, "unable to create session")
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
		renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// LogoutHandler clears the session cookie, invalidates its database row, and returns the user to the home page.
func LogoutHandler(db *sql.DB) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			renderHTTPError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		cookie, err := r.Cookie("session")
		if err != nil {
			renderHTTPError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		if !parseLimitedForm(w, r) {
			return
		}
		if !validCSRFToken(db, r, cookie.Value) {
			renderHTTPError(w, http.StatusForbidden, "invalid csrf token")
			return
		}

		if err := database.DeleteSession(db, cookie.Value); err != nil {
			renderHTTPError(w, http.StatusInternalServerError, "unable to log out")
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
