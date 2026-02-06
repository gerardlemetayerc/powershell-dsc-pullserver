package handlers

import (
	"net/http"
	"log"
	"fmt"
	"time"
	"database/sql"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/db"
	samlsp "github.com/crewjam/saml/samlsp"
	jwt "github.com/golang-jwt/jwt/v5"
)

// SAMLLoginHandler handles SAML login, user mapping, and JWT issuance
func SAMLLoginHandler(dbConn *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		attributesIface := samlsp.SessionFromContext(r.Context())
		appCfg, err := internal.LoadAppConfig("config.json")
		if err != nil {
			log.Printf("[REGISTER][CONFIG] Error loading config: %v", err)
			http.Error(w, "Server configuration error: unable to load config", http.StatusInternalServerError)
			return
		}
		//log.Printf("[SAML] SAML session attributes: %+v", attributesIface[])
		var email, firstName, lastName, role string
		if attributesIface == nil {
			log.Printf("[SAML] No SAML session attributes found in context")
			http.Error(w, "SAML session missing", http.StatusForbidden)
			return
		}
		attributes, ok := attributesIface.(samlsp.SessionWithAttributes)
		if !ok {
			log.Printf("[SAML] SessionFromContext is not SessionWithAttributes")
			http.Error(w, "SAML session invalid", http.StatusForbidden)
			return
		}
		attrMap := attributes.GetAttributes()
		if vals, ok := attrMap[appCfg.SAML.UserMapping.Email]; ok && len(vals) > 0 {
			email = vals[0]
		}
		if vals, ok := attrMap[appCfg.SAML.UserMapping.GivenName]; ok && len(vals) > 0 {
			firstName = vals[0]
		}
		if vals, ok := attrMap[appCfg.SAML.UserMapping.Sn]; ok && len(vals) > 0 {
			lastName = vals[0]
		}
		// SAML group extraction and role assignment
		groupAttr := appCfg.SAML.GroupMapping.Attribute
		adminValue := appCfg.SAML.GroupMapping.AdminValue
		userValue := appCfg.SAML.GroupMapping.UserValue
		isAdmin := false
		isUser := false
		if groupAttr != "" {
			if groupVals, ok := attrMap[groupAttr]; ok && len(groupVals) > 0 {
				for _, g := range groupVals {
					if g == adminValue {
						isAdmin = true
						log.Printf("[SAML] User is in admin group: %s", g)
					}
					if g == userValue {
						isUser = true
					}
				}
			}
		}
		if isAdmin {
			role = "admin"
		} else if isUser {
			role = "user"
		} else {
			http.Redirect(w, r, "/web/login?error=deny", http.StatusFound)
			return
		}

		if email == "" {
			log.Printf("[SAML] Aucun email trouvé dans les assertions SAML, accès refusé")
			http.Error(w, "Email SAML manquant", http.StatusForbidden)
			return
		}
		var userId int
		var isActiveBool bool
		err = dbConn.QueryRow("SELECT id, is_active FROM users WHERE email = ?", email).Scan(&userId, &isActiveBool)
		if err == sql.ErrNoRows {
			log.Printf("[SAML] Création nouvel utilisateur: %s %s <%s> (role: %s)", firstName, lastName, email, role)
			_, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active, last_logon_date, source, role) VALUES (?, ?, ?, '', 1, ?, 'saml', ?)", firstName, lastName, email, time.Now().Format("2006-01-02 15:04:05"), role)
			if err != nil {
				log.Printf("[SAML] Erreur création utilisateur: %v", err)
				http.Error(w, "Erreur création utilisateur", http.StatusInternalServerError)
				return
			}
		} else if err != nil && err != sql.ErrNoRows {
			log.Printf("[SAML] Erreur DB: %v", err)
			http.Error(w, "Erreur DB", http.StatusInternalServerError)
			return
		} else if !isActiveBool {
			// Compte inactif : refuse l'accès et redirige avec message
			log.Printf("[SAML] Compte inactif pour %s", email)
			http.Redirect(w, r, "/web/login?error=blocked", http.StatusFound)
			return
		} else {
			// Met à jour la date de dernière connexion
			if err := db.UpdateLastLogon(dbConn, userId); err != nil {
				log.Printf("[SAML] Erreur update last_logon_date: %v", err)
			}
			if err := db.UpdateRole(dbConn, userId, role); err != nil {
				log.Printf("[SAML] Erreur update role: %v", err)
			}
		}
		

	secret := []byte(appCfg.DSCPullServer.SharedAccessSecret)
	expiresAt := time.Now().Add(60 * time.Minute).Unix()
		claims := jwt.MapClaims{
			"sub": email,
			"exp": expiresAt,
			"role": role,
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
  document.cookie = 'jwt_token=' + %q + '; path=/; SameSite=Lax';
  window.location.replace('/web');
} catch(e) {
  window.location.replace('/web/login');
}
</script></body></html>`, signed, fmt.Sprintf("%d", expiresAt), signed)
	}
}
