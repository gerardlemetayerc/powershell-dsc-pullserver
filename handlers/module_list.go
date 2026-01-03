package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type ModuleInfo struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Version    string `json:"version"`
	Checksum   string `json:"checksum"`
	UploadedAt string `json:"uploaded_at"`
}

func ModuleListHandler(db *sql.DB) http.HandlerFunc {
	       return func(w http.ResponseWriter, r *http.Request) {
		       // Support de l'option ?count=1
		       if r.URL.Query().Get("count") == "1" {
			       var count int
			       err := db.QueryRow("SELECT COUNT(*) FROM modules").Scan(&count)
			       if err != nil {
				       w.WriteHeader(http.StatusInternalServerError)
				       w.Write([]byte("DB error"))
				       return
			       }
			       w.Header().Set("Content-Type", "application/json")
			       json.NewEncoder(w).Encode(map[string]int{"count": count})
			       return
		       }
		       rows, err := db.Query(`SELECT id, name, version, checksum, uploaded_at FROM modules ORDER BY uploaded_at DESC`)
		       if err != nil {
			       w.WriteHeader(http.StatusInternalServerError)
			       w.Write([]byte("DB error"))
			       return
		       }
		       defer rows.Close()
		       var modules []ModuleInfo
		       for rows.Next() {
			       var m ModuleInfo
			       if err := rows.Scan(&m.ID, &m.Name, &m.Version, &m.Checksum, &m.UploadedAt); err != nil {
				       continue
			       }
			       modules = append(modules, m)
		       }
		       w.Header().Set("Content-Type", "application/json")
		       json.NewEncoder(w).Encode(modules)
	       }
}
