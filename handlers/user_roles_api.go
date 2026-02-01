package handlers

import (
	"encoding/json"
	"net/http"
)

// UserRolesHandler retourne la liste des profils disponibles
func UserRolesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles := []string{"readonly", "admin"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(roles)
	}
}
