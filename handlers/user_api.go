package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"golang.org/x/crypto/bcrypt"
	"log"
	"time"
	"github.com/golang-jwt/jwt/v5"
	"go-dsc-pull/internal/schema"
	internaldb "go-dsc-pull/internal/db"
)


// Handler de login JWT
func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		type LoginResponse struct {
			Token string `json:"token"`
			ExpiresAt int64 `json:"expires_at"`
		}
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		row := db.QueryRow("SELECT id, password_hash FROM users WHERE email = ? AND is_active = 1", req.Username)
		var id int64
		var hash string
		err := row.Scan(&id, &hash)
		log.Printf("[LOGIN] Résultat SQL: id=%v, hash=%v, err=%v", id, hash, err)
		log.Printf("[LOGIN] Teste mot de passe: username='%s', password='%s', hash='%s'", req.Username, req.Password, hash)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
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
		secret := []byte("supersecretkey")
		expiresAt := time.Now().Add(60 * time.Minute).Unix()
		claims := jwt.MapClaims{
			"sub": req.Username,
			"exp": expiresAt,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		signed, err := token.SignedString(secret)
		if err != nil {
			http.Error(w, "Token error", http.StatusInternalServerError)
			return
		}
		resp := LoginResponse{Token: signed, ExpiresAt: expiresAt}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// List users
func ListUsersHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, first_name, last_name, email, is_active, created_at, last_logon_date FROM users")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		var users []schema.User
		for rows.Next() {
			var u schema.User
			rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.IsActive, &u.CreatedAt, &u.LastLogonDate)
			users = append(users, u)
		}
		json.NewEncoder(w).Encode(users)
	}
}

// Get user by id
func GetUserHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		log.Printf("[API] GetUserHandler id=%v", id)
		row := db.QueryRow("SELECT id, first_name, last_name, email, is_active, created_at, last_logon_date FROM users WHERE id = ?", id)
		var (
			u schema.User
			lastLogon sql.NullString
		)
		if err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.IsActive, &u.CreatedAt, &lastLogon); err != nil {
			log.Printf("[API] GetUserHandler SQL error: %v", err)
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
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
			   LastName string `json:"last_name"`
			   Email string `json:"email"`
			   Password string `json:"password"`
			   IsActive string `json:"is_active"`
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
		   res, err := db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, is_active) VALUES (?, ?, ?, ?, ?)", req.FirstName, req.LastName, req.Email, string(hash), isActive)
		   if err != nil {
			   http.Error(w, err.Error(), http.StatusInternalServerError)
			   return
		   }
		   id, _ := res.LastInsertId()
		   u := schema.User{
			   ID: id,
			   FirstName: req.FirstName,
			   LastName: req.LastName,
			   Email: req.Email,
			   IsActive: isActive,
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
		_, err := db.Exec("UPDATE users SET first_name=?, last_name=?, email=?, is_active=?, last_logon_date=? WHERE id=?", u.FirstName, u.LastName, u.Email, u.IsActive, u.LastLogonDate, id)
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
		_, err := db.Exec("DELETE FROM users WHERE id=?", id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		_, err := db.Exec("UPDATE users SET is_active=? WHERE id=?", isActive, id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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
		// Hash du mot de passe avec bcrypt
		hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Hash error", http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("UPDATE users SET password_hash=? WHERE id=?", string(hash), id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
