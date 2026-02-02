package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
)

// AgentByIdAPIHandler retourne les infos d'un agent donné
func AgentByIdAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
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

	       row := database.QueryRow(`SELECT agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter, registered_at, last_communication, has_error_last_report, state FROM agents WHERE agent_id = ?`, agentId)
	       var a schema.Agent
	       var hasErrorBool bool
	       var state *string
	       err = row.Scan(&a.AgentId, &a.NodeName, &a.LCMVersion, &a.RegistrationType, &a.CertificateThumbprint, &a.CertificateSubject, &a.CertificateIssuer, &a.CertificateNotBefore, &a.CertificateNotAfter, &a.RegisteredAt, &a.LastCommunication, &hasErrorBool, &state)
	       a.HasErrorLastReport = hasErrorBool
	       a.State = state
	       if err != nil {
		       http.Error(w, "Agent non trouvé", http.StatusNotFound)
		       return
	       }
	// Ajoute les configurations associées
	configs, err := db.GetAgentConfigurations(database, agentId)
	if err == nil {
		a.Configurations = configs
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(a)
}
