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
			       log.Println("[SAML] Liste complète des claims reçus :")
			       samlSession := samlsp.SessionFromContext(r.Context())
			       foundAttrs := false
			       if samlSession != nil {
				       if attrs, ok := samlSession.(interface{ GetAttributes() map[string][]string }); ok {
					       for k, vals := range attrs.GetAttributes() {
						       log.Printf("[SAML] %s = %v", k, vals)
						       foundAttrs = true
					       }
				       }
			       }
			       if !foundAttrs {
				       log.Println("[SAML] Aucun attribut SAML trouvé dans la session.")
			       }
			       // Toujours afficher les claims principaux pour debug
			       claimKeys := []string{
				       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/emailaddress",
				       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/givenname",
				       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/surname",
				       "http://schemas.xmlsoap.org/ws/2005/05/identity/claims/name",
				       "http://schemas.microsoft.com/identity/claims/displayname",
				       "http://schemas.microsoft.com/ws/2008/06/identity/claims/groups",
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
		       _, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active, last_logon_date) VALUES (?, ?, ?, '', 1, ?)", firstName, lastName, email, time.Now().Format("2006-01-02 15:04:05"))
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
	       } else {
		       // Met à jour la date de dernière connexion
		       if err := db.UpdateLastLogon(dbConn, userId); err != nil {
			       log.Printf("[SAML] Erreur update last_logon_date: %v", err)
		       }
	       }

	       // --- SAML group mapping for role assignment ---
	       appCfg, err := internal.LoadAppConfig("config.json")
	       if err != nil || appCfg.SAML.GroupMapping.Attribute == "" {
		       log.Printf("[SAML] Erreur lecture config group_mapping: %v", err)
		       http.Error(w, "Erreur config SAML group_mapping", http.StatusInternalServerError)
		       return
	       }
	       groupAttr := appCfg.SAML.GroupMapping.Attribute
	       groupValues := samlsp.AttributeFromContext(r.Context(), groupAttr)
	       role := "user" // default
	       if groupValues != "" {
		       // Support multi-value (comma or semicolon separated)
		       foundAdmin := false
		       foundUser := false
		       for _, v := range []string{groupValues} {
			       // If multiple values, split
			       for _, g := range splitGroups(v) {
				       if g == appCfg.SAML.GroupMapping.AdminValue {
					       foundAdmin = true
				       }
				       if g == appCfg.SAML.GroupMapping.UserValue {
					       foundUser = true
				       }
			       }
		       }
		       if foundAdmin {
			       role = "admin"
		       } else if foundUser {
			       role = "user"
		       }
	       }
	       // --- End SAML group mapping ---

	       // Update or create user with role
	       err = dbConn.QueryRow("SELECT id, is_active FROM users WHERE email = ?", email).Scan(&userId, &isActive)
	       if err == sql.ErrNoRows {
		       log.Printf("[SAML] Création nouvel utilisateur: %s %s <%s> (role: %s)", firstName, lastName, email, role)
		       _, err := dbConn.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active, last_logon_date, role) VALUES (?, ?, ?, '', 1, ?, ?)", firstName, lastName, email, time.Now().Format("2006-01-02 15:04:05"), role)
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
	       } else {
		       // Met à jour la date de dernière connexion et le rôle si changé
		       if err := db.UpdateLastLogon(dbConn, userId); err != nil {
			       log.Printf("[SAML] Erreur update last_logon_date: %v", err)
		       }
		       // Optionally update role if changed
		       _, err := dbConn.Exec("UPDATE users SET role=? WHERE id=?", role, userId)
		       if err != nil {
			       log.Printf("[SAML] Erreur update role: %v", err)
		       }
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
  document.cookie = 'jwt_token=' + %q + '; path=/; SameSite=Lax';
  window.location.replace('/web');
} catch(e) {
  window.location.replace('/web/login');
}
</script></body></html>`, signed, fmt.Sprintf("%d", expiresAt), signed)
       }
}

// Helper to split group values
func splitGroups(val string) []string {
       // Try comma, semicolon, or space
       var res []string
       for _, sep := range []string{",", ";", " "} {
	       if len(res) == 0 && len(val) > 0 && contains(val, sep) {
		       res = split(val, sep)
	       }
       }
       if len(res) == 0 && val != "" {
	       res = []string{val}
       }
       return res
}

func contains(s, sep string) bool { return len(sep) > 0 && len(s) > 0 && (len(split(s, sep)) > 1) }

func split(s, sep string) []string {
       var out []string
       for _, v := range []rune(s) {
	       if string(v) == sep {
		       out = append(out, "")
	       } else {
		       if len(out) == 0 {
			       out = append(out, "")
		       }
		       out[len(out)-1] += string(v)
	       }
       }
       return out
}
