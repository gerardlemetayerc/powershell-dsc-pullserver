package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// Génère un token API aléatoire (48 bytes, base64)
func GenerateAPIToken() (string, error) {
	tokenBytes := make([]byte, 48)
	_, err := rand.Read(tokenBytes)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(tokenBytes), nil
}

// Stocke le hash du token API pour un utilisateur
func StoreAPIToken(db *sql.DB, userId int64, plainToken, label string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainToken), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	_, err = db.Exec(`INSERT INTO user_api_tokens (user_id, token_hash, label, is_active, created_at) VALUES (?, ?, ?, 1, ?)`,
		userId, string(hash), label, time.Now().Format("2006-01-02 15:04:05"))
	return err
}

// Vérifie si un token API est valide et actif, retourne l'user_id si OK
func CheckAPIToken(db *sql.DB, token string) (int64, error) {
	rows, err := db.Query(`SELECT user_id, token_hash FROM user_api_tokens WHERE is_active=1 AND revoked_at IS NULL`)
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var userId int64
		var hash string
		if err := rows.Scan(&userId, &hash); err != nil {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)) == nil {
			return userId, nil
		}
	}
	return 0, errors.New("invalid token")
}
