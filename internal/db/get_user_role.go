package db

import (
	"database/sql"
)

// GetUserRole returns the user's ID and role by email, or an error if not found.
func GetUserRole(db *sql.DB, email string) (int64, string, error) {
	var id int64
	var role string
	row := db.QueryRow("SELECT id, role FROM users WHERE email = ?", email)
	if err := row.Scan(&id, &role); err != nil {
		return 0, "", err
	}
	return id, role, nil
}
