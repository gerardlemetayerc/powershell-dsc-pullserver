package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"go-dsc-pull/internal/buildinfo"
)

// AboutAPIHandler returns application version details from build info and database metadata.
func AboutAPIHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"build_version": buildinfo.Version,
		}

		var dbVersion sql.NullString
		err := db.QueryRow("SELECT db_version FROM dsc_infra_info WHERE id = 1").Scan(&dbVersion)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if dbVersion.Valid {
			response["db_version"] = dbVersion.String
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}
}