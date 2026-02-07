package utils

import (
	"database/sql"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

// IsAdmin vérifie si l'utilisateur courant (JWT) est admin
func IsAdmin(r *http.Request, dbConn *sql.DB) bool {
	   auth := r.Header.Get("Authorization")
	   if !strings.HasPrefix(auth, "Bearer ") {
		   return false
	   }
	   tokenStr := strings.TrimPrefix(auth, "Bearer ")
	   token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	   if err != nil {
		   return false
	   }
	   claims, ok := token.Claims.(jwt.MapClaims)
	   if !ok || claims["sub"] == nil {
		   return false
	   }
	   email := claims["sub"].(string)
	   row := dbConn.QueryRow("SELECT role FROM users WHERE email = ?", email)
	   var role string
	   log.Printf("[AUTH] Checking admin role for user %s (%s)", email, role)
	   if err := row.Scan(&role); err != nil {
		   return false
	   }
	   return role == "admin"
}
