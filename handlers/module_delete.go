package handlers

import (
	"database/sql"
	"go-dsc-pull/internal/auth"
	"net/http"
	"strconv"
)

// ModuleDeleteHandler supprime un module par id (admin seulement)
func ModuleDeleteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !auth.IsAdmin(r, db) {
			http.Error(w, "Forbidden: admin only", http.StatusForbidden)
			return
		}
		idStr := r.URL.Query().Get("id")
		id, err := strconv.Atoi(idStr)
		if err != nil || id <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid module id"))
			return
		}
		_, err = db.Exec(`DELETE FROM modules WHERE id = ?`, id)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("DB error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Module deleted"))
	}
}
