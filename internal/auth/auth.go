package auth

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"go-dsc-pull/internal"
	"github.com/golang-jwt/jwt/v5"
	"go-dsc-pull/internal/db"
)

// Middleware de vérification JWT Bearer
func JwtOrAPITokenAuthMiddleware(dbConn *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			   // Toujours définir auth
			   auth := r.Header.Get("Authorization")

			   // 1. Essaye de lire le JWT depuis le cookie
			   var tokenStr string
			   if cookie, err := r.Cookie("jwt_token"); err == nil {
				   tokenStr = cookie.Value
			   } else if strings.HasPrefix(auth, "Bearer ") {
				   // 2. Sinon, fallback sur le header Authorization (API, compat)
				   tokenStr = strings.TrimPrefix(auth, "Bearer ")
			   }

			   if tokenStr != "" {
				   appCfg, err := internal.LoadAppConfig("config.json")
				   if err != nil {
					   http.Error(w, "Server configuration error: unable to load config", http.StatusInternalServerError)
					   return
				   }
				   secret := []byte(appCfg.DSCPullServer.SharedAccessSecret)
				   jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
					   if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						   return nil, fmt.Errorf("Unexpected signing method")
					   }
					   return secret, nil
				   })
				   if err == nil && jwtToken.Valid {
					   if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
						   if sub, ok := claims["sub"].(string); ok {
							   ctx := context.WithValue(r.Context(), "userId", sub)
							   r = r.WithContext(ctx)
						   }
					   }
					   next.ServeHTTP(w, r)
					   return
				   }
				   http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
				   return
			   }

			   // Si ce n'est pas Bearer, tente API token
			   if strings.HasPrefix(auth, "Token ") {
				   tokenStr := strings.TrimPrefix(auth, "Token ")
				   userId, apiErr := db.CheckAPIToken(dbConn, tokenStr)
				   if apiErr == nil && userId > 0 {
					   // Vérifie que l'utilisateur existe et est actif
					   var isActiveBool bool
					   err := dbConn.QueryRow("SELECT is_active FROM users WHERE id = ?", userId).Scan(&isActiveBool)
					   if err != nil || !isActiveBool {
						   http.Error(w, "User not found or inactive", http.StatusUnauthorized)
						   return
					   }
					   ctx := context.WithValue(r.Context(), "userId", userId)
					   r = r.WithContext(ctx)
					   next.ServeHTTP(w, r)
					   return
				   }
				   http.Error(w, "Invalid or expired API token", http.StatusUnauthorized)
				   return
			   }
			   http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			   return
		})
	}
}
