package auth

import (
	"database/sql"
	"go-dsc-pull/internal/db"
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)

// IsAdmin vérifie si l'utilisateur courant (JWT) est admin
func IsAdmin(r *http.Request, dbConn *sql.DB) bool {
	   auth := r.Header.Get("Authorization")
	   var tokenStr string
	   if strings.HasPrefix(auth, "Bearer ") {
		   tokenStr = strings.TrimPrefix(auth, "Bearer ")
	   } else {
		   // Try cookies in priority order: jwt_token, jwt, token
		   cookie, err := r.Cookie("jwt_token")
		   if err == nil && cookie.Value != "" {
			   tokenStr = cookie.Value
		   } else {
			   cookie, err := r.Cookie("jwt")
			   if err == nil && cookie.Value != "" {
				   tokenStr = cookie.Value
			   } else {
				   cookie, err := r.Cookie("token")
				   if err == nil && cookie.Value != "" {
					   tokenStr = cookie.Value
				   } else {
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
		   return false
	   }

	   _, role, err := db.GetUserRole(dbConn, email)
	   if err != nil {
		   return false
	   }
	   return role == "admin"
}
