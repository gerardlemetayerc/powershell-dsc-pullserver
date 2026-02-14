package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/global"
)

// AuditListHandler retourne la liste des entrées d'audit (GET /api/v1/audit)
func AuditListHandler(w http.ResponseWriter, r *http.Request) {
	// Accès admin déjà vérifié par le middleware
	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB open error"))
		return
	}
	defer database.Close()

	rows, err := database.Query("SELECT id, user_email, action, target, details, ip_address, created_at FROM audit ORDER BY created_at DESC")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB query error"))
		return
	}
	defer rows.Close()

	audits := []schema.Audit{}
	for rows.Next() {
		var a schema.Audit
		if err := rows.Scan(&a.ID, &a.UserEmail, &a.Action, &a.Target, &a.Details, &a.IPAddress, &a.CreatedAt); err == nil {
			audits = append(audits, a)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audits)
}
