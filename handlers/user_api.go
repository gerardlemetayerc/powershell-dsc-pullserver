package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"log"
	"go-dsc-pull/internal/global"
	"time"
	"strconv"
	"github.com/golang-jwt/jwt/v5"
	internaldb "go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
)

// Retourne les infos de l'utilisateur courant (d'après le JWT)
func MyUserInfoHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Récupère l'email depuis le JWT (claim sub)
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["sub"] == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		email := claims["sub"].(string)
		row := db.QueryRow("SELECT id, first_name, last_name, email, role, is_active, created_at, last_logon_date FROM users WHERE email = ?", email)
		var u schema.User
		var lastLogon sql.NullString
		var isActiveBool bool
		if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Role, &isActiveBool, &u.CreatedAt, &lastLogon); err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		       u.IsActive = isActiveBool
		if lastLogon.Valid { u.LastLogonDate = &lastLogon.String } else { u.LastLogonDate = nil }
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(u)
	}
}

// Liste les tokens API d'un utilisateur
func ListUserAPITokensHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("id")
		rows, err := db.Query("SELECT id, user_id, label, is_active, created_at, revoked_at FROM user_api_tokens WHERE user_id = ?", userId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var tokens []schema.APIToken
		for rows.Next() {
			var t schema.APIToken
			var revokedAt sql.NullString
			rows.Scan(&t.ID, &t.UserID, &t.Label, &t.IsActive, &t.CreatedAt, &revokedAt)
			if revokedAt.Valid { t.RevokedAt = &revokedAt.String } else { t.RevokedAt = nil }
			tokens = append(tokens, t)
		}
		w.Header().Set("Content-Type", "application/json")
        // Toujours retourner un tableau (éventuellement vide)
        if tokens == nil {
            w.Write([]byte("[]"))
        } else {
            json.NewEncoder(w).Encode(tokens)
        }
	}
}

// Crée un nouveau token API pour un utilisateur (retourne le token plain)
func CreateUserAPITokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("id")
		var req struct{ Label string `json:"label"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		token, err := internaldb.GenerateAPIToken()
		if err != nil {
			http.Error(w, "Token generation error", http.StatusInternalServerError)
			return
		}
		// Stocke le hash
		id64, _ := strconv.ParseInt(userId, 10, 64)
		if err := internaldb.StoreAPIToken(db, id64, token, req.Label); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": token})
	}
}

// Révoque un token API (soft delete)
func RevokeUserAPITokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenId := r.PathValue("tokenid")
		_, err := db.Exec("UPDATE user_api_tokens SET is_active=0, revoked_at=? WHERE id=?", time.Now().Format("2006-01-02 15:04:05"), tokenId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Supprime un token API (hard delete)
func DeleteUserAPITokenHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenId := r.PathValue("tokenid")
		_, err := db.Exec("DELETE FROM user_api_tokens WHERE id=?", tokenId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler de login JWT
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		   // Plus besoin de LoginResponse, le token sera dans le cookie
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		       // On récupère aussi la source
			       row := db.QueryRow("SELECT id, password_hash, source, role FROM users WHERE email = ? AND is_active = 1", req.Username)
			       var id int64
			       var hash string
			       var source string
			       var role string
				       err := row.Scan(&id, &hash, &source, &role)
			       if err == sql.ErrNoRows {
				       http.Error(w, "Utilisateur ou mot de passe incorrect", http.StatusUnauthorized)
				       return
			       } else if err != nil {
				       log.Printf("[LOGIN] Erreur DB: %v", err)
				       http.Error(w, "Erreur interne", http.StatusInternalServerError)
				       return
			       }
		       // Si source = saml, on refuse l'authentification par mot de passe
		       if strings.ToLower(source) == "saml" {
			       http.Error(w, "Authentification par mot de passe interdite pour les utilisateurs SAML", http.StatusUnauthorized)
			       return
		       }
		       // Vérification du mot de passe
		       if bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
			       http.Error(w, "Unauthorized", http.StatusUnauthorized)
			       return
		       }
		// Met à jour la date de dernière connexion
		if err := internaldb.UpdateLastLogon(db, id); err != nil {
			log.Printf("[LOGIN] Erreur update last_logon_date: %v", err)
		}
		appCfg := global.AppConfig
		if appCfg == nil {
			log.Printf("[REGISTER][CONFIG] Error loading config: %v", err)
			http.Error(w, "Server configuration error: unable to load config", http.StatusInternalServerError)
			return
		}
		secret := []byte(appCfg.DSCPullServer.SharedAccessSecret)
		expiresAt := time.Now().Add(60 * time.Minute).Unix()
		   claims := jwt.MapClaims{
			   "sub": req.Username,
			   "exp": expiresAt,
			   "role": role,
		   }
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			http.Error(w, "Token error", http.StatusInternalServerError)
			return
		}
		   // Place le JWT dans un cookie HttpOnly/Secure/SameSite
		   http.SetCookie(w, &http.Cookie{
			   Name:     "jwt_token",
			   Value:    signed,
			   Path:     "/",
			   HttpOnly: true,
			   Secure:   true, // à adapter si tu testes en HTTP
			   SameSite: http.SameSiteStrictMode,
			   Expires:  time.Unix(expiresAt, 0),
		   })
		   w.Header().Set("Content-Type", "application/json")
		   json.NewEncoder(w).Encode(map[string]interface{}{"expires_at": expiresAt})
	}
}

// List users
func ListUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, first_name, last_name, email, role, is_active, created_at, last_logon_date, source FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var users []schema.User
		for rows.Next() {
			var u schema.User
			var lastLogon sql.NullString
			var isActiveBool bool
			if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Role, &isActiveBool, &u.CreatedAt, &lastLogon, &u.Source); err != nil {
				continue
			}
			u.IsActive = isActiveBool
			if lastLogon.Valid {
				u.LastLogonDate = &lastLogon.String
			} else {
				u.LastLogonDate = nil
			}
			users = append(users, u)
		}
		json.NewEncoder(w).Encode(users)
	}
}

// Get user by id
func GetUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		row := db.QueryRow("SELECT id, first_name, last_name, email, role, is_active, created_at, last_logon_date, source FROM users WHERE id = ?", id)
		var u schema.User
		var lastLogon sql.NullString
		var isActiveBool bool
		if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Role, &isActiveBool, &u.CreatedAt, &lastLogon, &u.Source); err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		u.IsActive = isActiveBool
		if lastLogon.Valid {
			u.LastLogonDate = &lastLogon.String
		} else {
			u.LastLogonDate = nil
		}
		json.NewEncoder(w).Encode(u)
	}
}

// Create user
func CreateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
			Email     string `json:"email"`
			Password  string `json:"password"`
			Role      string `json:"role"`
			IsActive  string `json:"is_active"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		if req.Password == "" {
			http.Error(w, "Le mot de passe est obligatoire", http.StatusBadRequest)
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Erreur hash mot de passe", http.StatusInternalServerError)
			return
		}
		isActive := req.IsActive == "1" || req.IsActive == "true"
		res, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, role, is_active) VALUES (?, ?, ?, ?, ?, ?)", req.FirstName, req.LastName, req.Email, string(hash), req.Role, isActive)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		u := schema.User{
			ID:        id,
			FirstName: req.FirstName,
			LastName:  req.LastName,
			Email:     req.Email,
			Role:      req.Role,
			IsActive:  isActive,
			Source:    "local",
		}
		json.NewEncoder(w).Encode(u)
	}
}

// Update user
func UpdateUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		var u schema.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		_, err := db.Exec("UPDATE users SET first_name=?, last_name=?, email=?, role=?, is_active=?, last_logon_date=? WHERE id=?", u.FirstName, u.LastName, u.Email, u.Role, u.IsActive, u.LastLogonDate, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

// Delete user
func DeleteUserHandler(db *sql.DB) http.HandlerFunc {
	       return func(w http.ResponseWriter, r *http.Request) {
		       id := r.PathValue("id")
		       // actorEmail = userId/sub from context
		       actorEmail := "?"
		       if r.Context().Value("userId") != nil {
			       if sub, ok := r.Context().Value("userId").(string); ok {
				       actorEmail = sub
			       }
		       }
		       // Get email of the user being deleted
		       deletedEmail := "?"
		       row := db.QueryRow("SELECT email FROM users WHERE id = ?", id)
		       var email string
		       if err := row.Scan(&email); err == nil {
			       deletedEmail = email
		       }
		       _, err := db.Exec("DELETE FROM users WHERE id=?", id)
		       if err != nil {
			       http.Error(w, err.Error(), http.StatusInternalServerError)
			       return
		       }
		       // Audit suppression
		       driverName := global.AppConfig.Database.Driver
		       _ = internaldb.InsertAudit(db, driverName, actorEmail, "delete", "user", "Deleted user: "+id+" ("+deletedEmail+")", "")
		       w.WriteHeader(http.StatusNoContent)
	       }
}

// Activate/Deactivate user
func SetUserActiveHandler(db *sql.DB) http.HandlerFunc {
	       return func(w http.ResponseWriter, r *http.Request) {
		       id := r.PathValue("id")
		       active := r.URL.Query().Get("active")
		       isActive := 0
		       if strings.ToLower(active) == "true" || active == "1" {
			       isActive = 1
		       }
		       // actorEmail = userId/sub from context
		       actorEmail := "?"
		       if r.Context().Value("userId") != nil {
			       if sub, ok := r.Context().Value("userId").(string); ok {
				       actorEmail = sub
			       }
		       }
		       // Get email of the user being activated/deactivated
		       targetEmail := "?"
		       row := db.QueryRow("SELECT email FROM users WHERE id = ?", id)
		       var email string
		       if err := row.Scan(&email); err == nil {
			       targetEmail = email
		       }
		       _, err := db.Exec("UPDATE users SET is_active=? WHERE id=?", isActive, id)
		       if err != nil {
			       http.Error(w, err.Error(), http.StatusInternalServerError)
			       return
		       }
		       // Audit activation/désactivation
		       driverName := global.AppConfig.Database.Driver
		       action := "deactivate"
		       if isActive == 1 {
			       action = "activate"
		       }
		       _ = internaldb.InsertAudit(db, driverName, actorEmail, action, "user", action+" user: "+id+" ("+targetEmail+")", "")
		       w.WriteHeader(http.StatusNoContent)
	       }
}

// Change user password
func ChangeUserPasswordHandler(db *sql.DB) http.HandlerFunc {
       return func(w http.ResponseWriter, r *http.Request) {
	       id := r.PathValue("id")
	       var req struct { NewPassword string `json:"new_password"` }
	       if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NewPassword == "" {
		       http.Error(w, "Bad request", http.StatusBadRequest)
		       return
	       }
	       hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	       if err != nil {
		       http.Error(w, "Hash error", http.StatusInternalServerError)
		       return
	       }
	       // Get email of the user performing the action (from context userId)
	       actorEmail := "?"
	       if r.Context().Value("userId") != nil {
		       if sub, ok := r.Context().Value("userId").(string); ok {
			       row := db.QueryRow("SELECT email FROM users WHERE id = ?", sub)
			       var email string
			       if err := row.Scan(&email); err == nil {
				       actorEmail = email
			       }
		       }
	       }
	       _, err = db.Exec("UPDATE users SET password_hash=? WHERE id=?", string(hash), id)
	       if err != nil {
		       http.Error(w, err.Error(), http.StatusInternalServerError)
		       return
	       }
	       // Audit changement de mot de passe
	       driverName := global.AppConfig.Database.Driver
	       _ = internaldb.InsertAudit(db, driverName, actorEmail, "update", "user", "Changed password for user: "+id, "")
	       w.WriteHeader(http.StatusNoContent)
       }
}
