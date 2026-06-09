package handlers

import (
	"crypto/subtle"
	"database/sql"
	"errors"
	"forum/database"
	"mime"
	"net/http"
	"unicode/utf8"
)

const (
	maxFormBodyBytes       int64 = 16 * 1024
	maxPostTitleChars            = 280
	maxPostContentChars          = 280
	maxCommentContentChars       = 280
	maxUsernameChars             = 280
)

func parseLimitedForm(w http.ResponseWriter, r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/x-www-form-urlencoded" {
		http.Error(w, "unsupported content type", http.StatusUnsupportedMediaType)
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxFormBodyBytes)

	if err := r.ParseForm(); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			http.Error(w, "request body too large", http.StatusRequestEntityTooLarge)
			return false
		}

		http.Error(w, "bad form data", http.StatusBadRequest)
		return false
	}

	return true
}

func exceedsCharacterLimit(value string, max int) bool {
	return utf8.RuneCountInString(value) > max
}

func validCSRFToken(db *sql.DB, r *http.Request, sessionID string) bool {
	expectedToken, err := database.GetCSRFTokenBySessionID(db, sessionID)
	if err != nil || expectedToken == "" {
		return false
	}

	submittedToken := r.FormValue("csrf_token")
	if submittedToken == "" || len(submittedToken) != len(expectedToken) {
		return false
	}

	return subtle.ConstantTimeCompare([]byte(submittedToken), []byte(expectedToken)) == 1
}
