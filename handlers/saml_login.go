package handlers

import (
	"net/http"
	"log"
	"fmt"
	"time"
	"database/sql"
	"go-dsc-pull/internal"
	samlsp "github.com/crewjam/saml/samlsp"
	jwt "github.com/golang-jwt/jwt/v5"
)

// SAMLLoginHandler handles SAML login, user mapping, and JWT issuance
func SAMLLoginHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("[SAML] Assertions reçues (toutes clés) :")
		claimKeys := []string{
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
			"http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
			"http://schemas.microsoft.com/identity/claims/displayname",
		}
		for _, k := range claimKeys {
			v := samlsp.AttributeFromContext(r.Context(), k)
			log.Printf("[SAML] %s = %s", k, v)
		}

		mapping, err := internal.GetSAMLUserMapping()
		if err != nil {
			log.Printf("[SAML] Erreur lecture mapping SAML: %v", err)
			http.Error(w, "Erreur mapping SAML", http.StatusInternalServerError)
			return
		}
		getAttr := func(uri string) string {
			return samlsp.AttributeFromContext(r.Context(), uri)
		}
		email := getAttr(mapping["email"])
		firstName := getAttr(mapping["givenName"])
		lastName := getAttr(mapping["sn"])
		if email == "" {
			log.Printf("[SAML] Aucun email trouvé dans les assertions SAML, accès refusé")
			http.Error(w, "Email SAML manquant", http.StatusForbidden)
			return
		}
		       var userId int
		       var isActive int
		       err = dbConn.QueryRow("SELECT id, is_active FROM users WHERE email = ?", email).Scan(&userId, &isActive)
		       if err == sql.ErrNoRows {
			       log.Printf("[SAML] Création nouvel utilisateur: %s %s <%s>", firstName, lastName, email)
			       _, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, '', 1)", firstName, lastName, email)
			       if err != nil {
				       log.Printf("[SAML] Erreur création utilisateur: %v", err)
				       http.Error(w, "Erreur création utilisateur", http.StatusInternalServerError)
				       return
			       }
		       } else if err != nil && err != sql.ErrNoRows {
			       log.Printf("[SAML] Erreur DB: %v", err)
			       http.Error(w, "Erreur DB", http.StatusInternalServerError)
			       return
					} else if isActive == 0 {
						// Compte inactif : refuse l'accès et redirige avec message
						log.Printf("[SAML] Compte inactif pour %s", email)
						http.Redirect(w, r, "/web/login?error=blocked", http.StatusFound)
						return
					}
		secret := []byte("supersecretkey")
		expiresAt := time.Now().Add(60 * time.Minute).Unix()
		claims := jwt.MapClaims{
			"sub": email,
			"exp": expiresAt,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			log.Printf("[SAML] Erreur génération JWT: %v", err)
			http.Error(w, "Erreur JWT", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, `<!DOCTYPE html><html lang='fr'><head><meta charset='UTF-8'><title>Connexion SAML...</title></head><body>Connexion en cours...<script>
try {
  localStorage.setItem('jwt_token', %q);
  localStorage.setItem('jwt_exp', %q);
  window.location.replace('/web');
} catch(e) {
  window.location.replace('/web/login');
}
</script></body></html>`, signed, fmt.Sprintf("%d", expiresAt))
	}
}
