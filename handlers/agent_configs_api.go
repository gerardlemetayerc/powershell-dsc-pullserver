package handlers

import (
	"encoding/json"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/schema"
	"log"
	"net/http"
	"time"
)

// AgentConfigsAPIHandler retourne la liste des configurations pour un agent donné
func AgentConfigsAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	agentId := r.PathValue("id")
	if agentId == "" {
		http.Error(w, "AgentId manquant", http.StatusBadRequest)
		return
	}
	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		log.Printf("[API][DB] Erreur ouverture DB: %v", err)
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	bindings, err := db.GetAgentConfigurationBindings(database, agentId)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	decorateBindingsWithNextExecution(bindings, time.Now().UTC())

	w.Header().Set("Content-Type", "application/json")
	if bindings == nil {
		bindings = make([]schema.AgentConfigurationBinding, 0)
	}
	_ = json.NewEncoder(w).Encode(bindings)
}
