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

	       // Support de l'option ?count=1
	       if r.URL.Query().Get("count") == "1" {
		       var count int
		       err := database.QueryRow("SELECT COUNT(*) FROM agents").Scan(&count)
		       if err != nil {
			       http.Error(w, "DB count error", http.StatusInternalServerError)
			       return
		       }
		       w.Header().Set("Content-Type", "application/json")
		       _ = json.NewEncoder(w).Encode(map[string]int{"count": count})
		       return
	       }

		       // Filtrage dynamique
		       q := `SELECT agent_id, node_name, lcm_version, registration_type, certificate_thumbprint, certificate_subject, certificate_issuer, certificate_notbefore, certificate_notafter, registered_at, last_communication, has_error_last_report FROM agents WHERE 1=1`
		       args := []interface{}{}
		       nodeName := r.URL.Query().Get("node_name")
		       if nodeName != "" {
			       q += " AND node_name = ?"
			       args = append(args, nodeName)
		       }
		       hasError := r.URL.Query().Get("has_error_last_report")
		       if hasError == "true" {
			       q += " AND has_error_last_report = 1"
		       } else if hasError == "false" {
			       q += " AND has_error_last_report = 0"
		       }
		       rows, err := database.Query(q, args...)
		       if err != nil {
			       http.Error(w, "DB query error", http.StatusInternalServerError)
			       return
		       }
		       defer rows.Close()

		       agents := []schema.Agent{}
		       for rows.Next() {
			       var a schema.Agent
			       var hasErrorInt int
			       if err := rows.Scan(&a.AgentId, &a.NodeName, &a.LCMVersion, &a.RegistrationType, &a.CertificateThumbprint, &a.CertificateSubject, &a.CertificateIssuer, &a.CertificateNotBefore, &a.CertificateNotAfter, &a.RegisteredAt, &a.LastCommunication, &hasErrorInt); err == nil {
				       a.HasErrorLastReport = hasErrorInt != 0
				       agents = append(agents, a)
			       }
		       }
		       w.Header().Set("Content-Type", "application/json")
		       _ = json.NewEncoder(w).Encode(agents)
}
