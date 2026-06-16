package auth

import (
	"database/sql"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"net/http"
	"os"
	"strconv"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)

// IsAdmin vérifie si l'utilisateur courant (JWT) est admin
func IsAdmin(r *http.Request, dbConn *sql.DB) bool {
	// First, trust authenticated middleware context (JWT or API token path).
	if ctxUser := r.Context().Value("userId"); ctxUser != nil {
		switch v := ctxUser.(type) {
		case int64:
			return isAdminByID(dbConn, v)
		case int:
			return isAdminByID(dbConn, int64(v))
		case string:
			if v == "" {
				return false
			}
			if id, err := strconv.ParseInt(v, 10, 64); err == nil {
				return isAdminByID(dbConn, id)
			}
			_, role, err := db.GetUserRole(dbConn, v)
			return err == nil && role == "admin"
		}
	}

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
	secretValue := os.Getenv("JWT_SECRET")
	if secretValue == "" {
		secretValue = ResolveJWTSecret(appCfg)
	}
	if secretValue == "" {
		return false
	}
	secret := []byte(secretValue)

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

func isAdminByID(dbConn *sql.DB, userID int64) bool {
	if userID <= 0 {
		return false
	}
	var role string
	if err := dbConn.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role); err != nil {
		return false
	}
	return role == "admin"
}
