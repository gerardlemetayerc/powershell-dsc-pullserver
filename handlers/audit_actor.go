package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

// resolveAuditActor returns a stable actor identifier for audit logs.
// It supports JWT contexts (email string) and API-token contexts (numeric user id).
func resolveAuditActor(dbConn *sql.DB, r *http.Request) string {
	ctxUser := r.Context().Value("userId")
	if ctxUser == nil {
		return "?"
	}

	switch v := ctxUser.(type) {
	case string:
		if v == "" {
			return "?"
		}
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			return lookupUserEmailByID(dbConn, id)
		}
		return v
	case int64:
		return lookupUserEmailByID(dbConn, v)
	case int:
		return lookupUserEmailByID(dbConn, int64(v))
	default:
		return "?"
	}
}

func lookupUserEmailByID(dbConn *sql.DB, userID int64) string {
	if userID <= 0 {
		return "?"
	}
	if dbConn == nil {
		return strconv.FormatInt(userID, 10)
	}
	var email string
	if err := dbConn.QueryRow("SELECT email FROM users WHERE id = ?", userID).Scan(&email); err != nil || email == "" {
		return strconv.FormatInt(userID, 10)
	}
	return email
}
