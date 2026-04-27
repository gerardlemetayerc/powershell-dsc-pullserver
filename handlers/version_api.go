package handlers

import (
	"encoding/json"
	"go-dsc-pull/internal/global"
	"net/http"
)

func VersionAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"version":           global.AppVersion,
		"target_db_version": global.TargetDBVersion,
	})
}
