package db

import (
	"database/sql"
	"fmt"
)

// Role constants
const (
	RoleAdministrator = "Administrator"
	RoleReadOnly      = "Read-Only"
)

// GetUserRoles returns all role names for a user
func GetUserRoles(db *sql.DB, userID int64) ([]string, error) {
	query := `
		SELECT r.name 
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ?
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, nil
}

// GetUserRolesByEmail returns all role names for a user by email
func GetUserRolesByEmail(db *sql.DB, email string) ([]string, error) {
	query := `
		SELECT r.name 
		FROM roles r
		INNER JOIN user_roles ur ON r.id = ur.role_id
		INNER JOIN users u ON u.id = ur.user_id
		WHERE u.email = ?
	`
	rows, err := db.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var roleName string
		if err := rows.Scan(&roleName); err != nil {
			return nil, err
		}
		roles = append(roles, roleName)
	}
	return roles, nil
}

// HasRole checks if a user has a specific role
func HasRole(db *sql.DB, userID int64, roleName string) (bool, error) {
	roles, err := GetUserRoles(db, userID)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == roleName {
			return true, nil
		}
	}
	return false, nil
}

// HasRoleByEmail checks if a user has a specific role by email
func HasRoleByEmail(db *sql.DB, email string, roleName string) (bool, error) {
	roles, err := GetUserRolesByEmail(db, email)
	if err != nil {
		return false, err
	}
	for _, r := range roles {
		if r == roleName {
			return true, nil
		}
	}
	return false, nil
}

// IsAdministrator checks if a user has the Administrator role
func IsAdministrator(db *sql.DB, userID int64) (bool, error) {
	return HasRole(db, userID, RoleAdministrator)
}

// IsAdministratorByEmail checks if a user has the Administrator role by email
func IsAdministratorByEmail(db *sql.DB, email string) (bool, error) {
	return HasRoleByEmail(db, email, RoleAdministrator)
}

// AssignRole assigns a role to a user
func AssignRole(db *sql.DB, userID int64, roleName string) error {
	// Get role ID by name
	var roleID int64
	err := db.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("role not found: %s", roleName)
	}

	// Assign role to user
	_, err = db.Exec("INSERT OR IGNORE INTO user_roles (user_id, role_id) VALUES (?, ?)", userID, roleID)
	return err
}

// RemoveRole removes a role from a user
func RemoveRole(db *sql.DB, userID int64, roleName string) error {
	// Get role ID by name
	var roleID int64
	err := db.QueryRow("SELECT id FROM roles WHERE name = ?", roleName).Scan(&roleID)
	if err != nil {
		return fmt.Errorf("role not found: %s", roleName)
	}

	// Remove role from user
	_, err = db.Exec("DELETE FROM user_roles WHERE user_id = ? AND role_id = ?", userID, roleID)
	return err
}

// GetAllRoles returns all available roles
func GetAllRoles(db *sql.DB) ([]map[string]interface{}, error) {
	rows, err := db.Query("SELECT id, name, description, created_at FROM roles")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []map[string]interface{}
	for rows.Next() {
		var id int64
		var name, description, createdAt string
		if err := rows.Scan(&id, &name, &description, &createdAt); err != nil {
			return nil, err
		}
		roles = append(roles, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": description,
			"created_at":  createdAt,
		})
	}
	return roles, nil
}
