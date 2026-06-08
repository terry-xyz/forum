package handlers

import (
	"crypto/subtle"
	"database/sql"
	"forum/database"
	"net/http"
)

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
