package auth

import (
	"database/sql"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
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

	appCfg := global.AppConfig
	if appCfg == nil {
		return false
	}
	secret := []byte(appCfg.DSCPullServer.SharedAccessSecret)

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
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
