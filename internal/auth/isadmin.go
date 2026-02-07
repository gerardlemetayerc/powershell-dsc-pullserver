package auth

import (
	"database/sql"
	"go-dsc-pull/internal/db"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"log"
)

// IsAdmin vérifie si l'utilisateur courant (JWT) est admin
func IsAdmin(r *http.Request, dbConn *sql.DB) bool {
	   auth := r.Header.Get("Authorization")
	   var tokenStr string
	   log.Printf("[AUTH] Checking admin role for token: %s", auth)
	   if strings.HasPrefix(auth, "Bearer ") {
		   tokenStr = strings.TrimPrefix(auth, "Bearer ")
		   log.Printf("[AUTH] Using Authorization header token")
	   } else {
		   // Try cookies in priority order: jwt_token, jwt, token
		   cookie, err := r.Cookie("jwt_token")
		   if err == nil && cookie.Value != "" {
			   tokenStr = cookie.Value
			   log.Printf("[AUTH] Using jwt_token cookie")
		   } else {
			   cookie, err := r.Cookie("jwt")
			   if err == nil && cookie.Value != "" {
				   tokenStr = cookie.Value
				   log.Printf("[AUTH] Using jwt cookie")
			   } else {
				   cookie, err := r.Cookie("token")
				   if err == nil && cookie.Value != "" {
					   tokenStr = cookie.Value
					   log.Printf("[AUTH] Using token cookie")
				   } else {
					   log.Printf("[AUTH] No valid JWT found in header or cookies")
					   return false
				   }
			   }
		   }
	   }

	   token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	   if err != nil {
		   return false
	   }
	   claims, ok := token.Claims.(jwt.MapClaims)
	   if !ok || claims["sub"] == nil {
		   return false
	   }
	   emailRaw := claims["sub"]
	   email, ok := emailRaw.(string)
	   if !ok || email == "" {
		   log.Printf("[AUTH] Invalid or missing email claim: %v", emailRaw)
		   return false
	   }

	   id, role, err := db.GetUserRole(dbConn, email)
	   log.Printf("[AUTH] Checking admin role for user %s (id=%d, role=%s)", email, id, role)
	   if err != nil {
		   return false
	   }
	   return role == "admin"
}
