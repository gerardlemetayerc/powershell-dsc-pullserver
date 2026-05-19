package handlers

import (
	"net/http"
	"time"
)

func clearCookie(w http.ResponseWriter, name string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	})
}

func setJWTCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	clearCookie(w, "token")
	http.SetCookie(w, &http.Cookie{
		Name:     "jwt_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Expires:  expiresAt,
	})
}

func clearJWTCookie(w http.ResponseWriter) {
	clearCookie(w, "jwt_token")
	clearCookie(w, "token")
	clearCookie(w, "saml_session")
}