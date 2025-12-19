package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
)

// AgentAPIHandler retourne la liste des agents (table agents)
func AgentAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
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

	rows, err := database.Query(`SELECT agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter, registered_at FROM agents`)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	agents := []schema.Agent{}
	for rows.Next() {
		var a schema.Agent
		if err := rows.Scan(&a.AgentId, &a.NodeName, &a.LCMVersion, &a.RegistrationType, &a.CertificateThumbprint, &a.CertificateSubject, &a.CertificateIssuer, &a.CertificateNotBefore, &a.CertificateNotAfter, &a.RegisteredAt); err == nil {
			agents = append(agents, a)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agents)
}
