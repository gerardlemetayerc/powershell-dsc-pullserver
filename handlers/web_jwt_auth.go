package handlers

import (
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"fmt"
	"os"
)

func StatFile(path string) (os.FileInfo, error) {
	return os.Stat(path)
}


// WebJWTAuthMiddleware protège les routes web en vérifiant le cookie 'token' (JWT)
func WebJWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		// 1. Cherche le header Authorization Bearer
		   auth := r.Header.Get("Authorization")
		   if strings.HasPrefix(auth, "Bearer ") {
			   tokenStr = strings.TrimPrefix(auth, "Bearer ")
		   } else {
			   // 2. Sinon, cherche le cookie 'jwt_token'
			   cookie, err := r.Cookie("jwt_token")
			   if err == nil {
				   tokenStr = cookie.Value
			   } else {
				   fmt.Printf("[WebJWTAuthMiddleware] Pas de cookie 'token'\n")
			   }
		   }
		if tokenStr == "" {
			// Redirige vers login si pas de token
			http.Redirect(w, r, "/web/login", http.StatusFound)
			return
		}
		// 3. Valide le JWT (clé à adapter selon ton projet)
		secret := []byte("supersecretkey")
		jwtToken, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !jwtToken.Valid {
			http.Redirect(w, r, "/web/login", http.StatusFound)
			return
		}
		// Token valide, continue
		next.ServeHTTP(w, r)
	})
}
