package handlers

import (
	"database/sql"
	"go-dsc-pull/internal/auth"
	"net/http"
	"strconv"
	"fmt"
	"go-dsc-pull/internal/global"
	internaldb "go-dsc-pull/internal/db"
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
		   // Récupère les infos du module avant suppression
		   var modName, modVersion string
		   err = db.QueryRow(`SELECT name, version FROM modules WHERE id = ?`, id).Scan(&modName, &modVersion)
		   if err != nil {
			   w.WriteHeader(http.StatusInternalServerError)
			   w.Write([]byte("DB error: unable to fetch module info"))
			   return
		   }
		   _, err = db.Exec(`DELETE FROM modules WHERE id = ?`, id)
		   if err != nil {
			   w.WriteHeader(http.StatusInternalServerError)
			   w.Write([]byte("DB error"))
			   return
		   }
		   // Audit suppression
		   driverName := global.AppConfig.Database.Driver
		   user := "?"
		   if r.Context().Value("userId") != nil {
			   if sub, ok := r.Context().Value("userId").(string); ok {
				   user = sub
			   }
			   _ = internaldb.InsertAudit(db, driverName, user, "delete", "module", fmt.Sprintf("Deleted module: %s v%s", modName, modVersion), "")
		   }
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Module deleted"))
	}
}