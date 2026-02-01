package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"crypto/sha1"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/logs"
)

// ptr retourne un pointeur sur une chaîne (utilitaire)
func ptr(s string) *string { return &s }


// PreEnrollAgentHandler allows pre-registration of an agent with only NodeName
func PreEnrollAgentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		NodeName string `json:"node_name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.NodeName == "" {
		http.Error(w, "Invalid or missing node_name", http.StatusBadRequest)
		return
	}
	// Generate a temporary AgentId (SHA1 of NodeName + 'TEMP')
	h := sha1.New()
	h.Write([]byte(req.NodeName + "TEMP"))
	tempAgentId := fmt.Sprintf("TEMP-%x", h.Sum(nil)[:8])

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

	// Insert agent (ignore if already exists)
	_, err = database.Exec(`INSERT OR IGNORE INTO agents (agent_id, node_name, last_communication, state) VALUES (?, ?, ?, ?)`, tempAgentId, req.NodeName, "0000-00-01 00:00:00", "waiting_for_registration")
	if err != nil {
		log.Printf("[API][DB] Erreur insertion agent: %v", err)
		_ = logs.WriteLogFile(fmt.Sprintf("ERROR [API][PRE-ENROLL] Erreur insertion agent: %v (NodeName=%s)", err, req.NodeName))
		http.Error(w, "DB insert error", http.StatusInternalServerError)
		return
	}

	_ = logs.WriteLogFile(fmt.Sprintf("INFO [API][PRE-ENROLL] Agent pré-enregistré: NodeName=%s, AgentId=%s", req.NodeName, tempAgentId))

	agent := schema.Agent{
		AgentId:           tempAgentId,
		NodeName:          req.NodeName,
		LastCommunication: "0000-00-01 00:00:00",
		State:             ptr("waiting_for_registration"),
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agent)
}
