package db

import (
	"database/sql"
	"testing"

	_ "modernc.org/sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	// Create in-memory SQLite database for testing
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create schema
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		is_active BOOLEAN DEFAULT 1,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_logon_date TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS roles (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	INSERT INTO roles (id, name, description) VALUES 
		(1, 'Administrator', 'Full access to all server features'),
		(2, 'Read-Only', 'Limited to viewing configurations and status');

	CREATE TABLE IF NOT EXISTS user_roles (
		user_id INTEGER NOT NULL,
		role_id INTEGER NOT NULL,
		assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (user_id, role_id),
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
	);
	`

	_, err = db.Exec(schema)
	if err != nil {
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db
}

func TestAssignRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	result, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)",
		"Test", "User", "test@example.com", "hash", 1)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, _ := result.LastInsertId()

	// Test assigning Administrator role
	err = AssignRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Errorf("Failed to assign Administrator role: %v", err)
	}

	// Verify role was assigned
	hasRole, err := HasRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Errorf("Failed to check role: %v", err)
	}
	if !hasRole {
		t.Error("User should have Administrator role but doesn't")
	}

	// Test assigning Read-Only role
	err = AssignRole(db, userID, RoleReadOnly)
	if err != nil {
		t.Errorf("Failed to assign Read-Only role: %v", err)
	}

	// Verify both roles are assigned
	roles, err := GetUserRoles(db, userID)
	if err != nil {
		t.Errorf("Failed to get user roles: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}
}

func TestRemoveRole(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	result, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)",
		"Test", "User", "test@example.com", "hash", 1)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, _ := result.LastInsertId()

	// Assign a role
	err = AssignRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Remove the role
	err = RemoveRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Errorf("Failed to remove role: %v", err)
	}

	// Verify role was removed
	hasRole, err := HasRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Errorf("Failed to check role: %v", err)
	}
	if hasRole {
		t.Error("User should not have Administrator role but does")
	}
}

func TestIsAdministrator(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	result, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)",
		"Admin", "User", "admin@example.com", "hash", 1)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, _ := result.LastInsertId()

	// Initially should not be admin
	isAdmin, err := IsAdministrator(db, userID)
	if err != nil {
		t.Errorf("Failed to check admin status: %v", err)
	}
	if isAdmin {
		t.Error("User should not be admin initially")
	}

	// Assign Administrator role
	err = AssignRole(db, userID, RoleAdministrator)
	if err != nil {
		t.Fatalf("Failed to assign Administrator role: %v", err)
	}

	// Now should be admin
	isAdmin, err = IsAdministrator(db, userID)
	if err != nil {
		t.Errorf("Failed to check admin status: %v", err)
	}
	if !isAdmin {
		t.Error("User should be admin after role assignment")
	}
}

func TestGetUserRolesByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create a test user
	result, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)",
		"Test", "User", "test@example.com", "hash", 1)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	userID, _ := result.LastInsertId()

	// Assign Read-Only role
	err = AssignRole(db, userID, RoleReadOnly)
	if err != nil {
		t.Fatalf("Failed to assign role: %v", err)
	}

	// Get roles by email
	roles, err := GetUserRolesByEmail(db, "test@example.com")
	if err != nil {
		t.Errorf("Failed to get user roles by email: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
	if len(roles) > 0 && roles[0] != RoleReadOnly {
		t.Errorf("Expected Read-Only role, got %s", roles[0])
	}
}

func TestGetAllRoles(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	roles, err := GetAllRoles(db)
	if err != nil {
		t.Errorf("Failed to get all roles: %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("Expected 2 roles, got %d", len(roles))
	}

	// Verify role names
	roleNames := make(map[string]bool)
	for _, role := range roles {
		roleNames[role["name"].(string)] = true
	}
	if !roleNames[RoleAdministrator] || !roleNames[RoleReadOnly] {
		t.Error("Expected both Administrator and Read-Only roles")
	}
}
