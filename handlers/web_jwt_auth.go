package handlers

import (
	"net/http"
	"strings"
	"github.com/golang-jwt/jwt/v5"
	"go-dsc-pull/internal/auth"
	"go-dsc-pull/internal/global"
	"fmt"
)


// WebJWTAuthMiddleware protège les routes web en vérifiant le cookie 'token' (JWT)
func WebJWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		// 1. Cherche le header Authorization Bearer
		   authHeader := r.Header.Get("Authorization")
		   if strings.HasPrefix(authHeader, "Bearer ") {
			   tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
		   } else {
			   // 2. Sinon, cherche le cookie 'jwt_token'
			   cookie, err := r.Cookie("jwt_token")
			   if err == nil {
				   tokenStr = cookie.Value
			   } else {
				   fmt.Printf("[WebJWTAuthMiddleware] Pas de cookie 'jwt_token'\n")
			   }
		   }
		if tokenStr == "" {
			// Redirige vers login si pas de token
			http.Redirect(w, r, "/web/login", http.StatusFound)
			return
		}
		// 3. Valide le JWT et extrait les claims
		appCfg := global.AppConfig
		if appCfg == nil {
			http.Error(w, "Server configuration error: unable to load config", http.StatusInternalServerError)
			return
		}
		secretValue := auth.ResolveJWTSecret(appCfg)
		if secretValue == "" {
			http.Error(w, "Server configuration error: missing shared_secret", http.StatusInternalServerError)
			return
		}
		secret := []byte(secretValue)
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
