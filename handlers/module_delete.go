package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
)

func ModuleDeleteHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
