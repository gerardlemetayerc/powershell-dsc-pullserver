package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"go-dsc-pull/internal/db"
)

// AgentConfigsAPIHandlerPostDelete gère POST (ajout) et DELETE (suppression) sur /api/v1/agents/{id}/configs
func AgentConfigsAPIHandlerPostDelete(w http.ResponseWriter, r *http.Request) {
	agentId := r.PathValue("id")
	if agentId == "" {
		http.Error(w, "AgentId manquant", http.StatusBadRequest)
		return
	}
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		log.Printf("[API][DB] Erreur chargement config DB: %v", err)
		http.Error(w, "DB config error", http.StatusInternalServerError)
		return
	}
	database, err := db.OpenDB(dbCfg)
	if err != nil {
		log.Printf("[API][DB] Erreur ouverture DB: %v", err)
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	switch r.Method {
	case http.MethodPost:
		var req struct { ConfigurationName string `json:"configuration_name"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ConfigurationName == "" {
			http.Error(w, "Nom de configuration manquant ou invalide", http.StatusBadRequest)
			return
		}
		_, err := database.Exec(`INSERT OR REPLACE INTO agent_configurations (agent_id, configuration_name) VALUES (?, ?)`, agentId, req.ConfigurationName)
		if err != nil {
			http.Error(w, "Erreur insertion config", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	case http.MethodDelete:
		var req struct { ConfigurationName string `json:"configuration_name"` }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ConfigurationName == "" {
			http.Error(w, "Nom de configuration manquant ou invalide", http.StatusBadRequest)
			return
		}
		_, err := database.Exec(`DELETE FROM agent_configurations WHERE agent_id = ? AND configuration_name = ?`, agentId, req.ConfigurationName)
		if err != nil {
			http.Error(w, "Erreur suppression config", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
