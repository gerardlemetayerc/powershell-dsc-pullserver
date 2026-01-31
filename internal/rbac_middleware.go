package internal

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"go-dsc-pull/internal/db"
)

// RBACMiddleware checks if the user has the required role(s) to access the endpoint
func RBACMiddleware(dbConn *sql.DB, requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user email from context (set by JWT middleware)
			userEmail := ""
			
			// Try to get from context first (API token auth)
			if userId := r.Context().Value("userId"); userId != nil {
				// If userId is a string, it's the email from JWT
				if email, ok := userId.(string); ok {
					userEmail = email
				} else if id, ok := userId.(int64); ok {
					// If it's an int64, it's from API token, need to lookup email
					var email string
					err := dbConn.QueryRow("SELECT email FROM users WHERE id = ?", id).Scan(&email)
					if err != nil {
						log.Printf("[RBAC] Error getting email for user ID %d: %v", id, err)
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return
					}
					userEmail = email
				}
			}

			// If not in context, try to extract from JWT token
			if userEmail == "" {
				auth := r.Header.Get("Authorization")
				if strings.HasPrefix(auth, "Bearer ") {
					tokenStr := strings.TrimPrefix(auth, "Bearer ")
					token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
					if err == nil {
						if claims, ok := token.Claims.(jwt.MapClaims); ok {
							if sub, ok := claims["sub"].(string); ok {
								userEmail = sub
							}
						}
					}
				}
			}

			if userEmail == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Get user roles
			userRoles, err := db.GetUserRolesByEmail(dbConn, userEmail)
			if err != nil {
				log.Printf("[RBAC] Error getting roles for user %s: %v", userEmail, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Convert user roles to a map for O(1) lookup
			userRolesMap := make(map[string]bool)
			for _, role := range userRoles {
				userRolesMap[role] = true
			}

			// Check if user has any of the required roles
			hasRequiredRole := false
			for _, requiredRole := range requiredRoles {
				if userRolesMap[requiredRole] {
					hasRequiredRole = true
					break
				}
			}

			if !hasRequiredRole {
				log.Printf("[RBAC] User %s does not have required role(s): %v (user has: %v)", userEmail, requiredRoles, userRoles)
				http.Error(w, "Forbidden: insufficient permissions", http.StatusForbidden)
				return
			}

			// Add roles to context for handlers to use if needed
			ctx := context.WithValue(r.Context(), "userRoles", userRoles)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AdminOnlyMiddleware is a convenience wrapper for Administrator-only endpoints
func AdminOnlyMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
	return RBACMiddleware(dbConn, db.RoleAdministrator)
}

// AuthenticatedMiddleware allows both Administrator and Read-Only users
func AuthenticatedMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
	return RBACMiddleware(dbConn, db.RoleAdministrator, db.RoleReadOnly)
}
