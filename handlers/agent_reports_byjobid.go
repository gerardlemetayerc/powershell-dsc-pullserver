package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"log"
)

// AgentReportsByJobIdHandler retourne un rapport précis d'un agent par JobId
func AgentReportsByJobIdHandler(w http.ResponseWriter, r *http.Request) {
	agentId := r.PathValue("id")
	jobId := r.PathValue("jobid")
	if agentId == "" || jobId == "" {
		http.Error(w, "AgentId ou JobId manquant", http.StatusBadRequest)
		return
	}
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		http.Error(w, "DB config error", http.StatusInternalServerError)
		return
	}
	database, err := db.OpenDB(dbCfg)
	if err != nil {
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	var query string
	switch dbCfg.Driver {
	case "sqlite":
		query = `SELECT raw_json FROM reports WHERE agent_id = ? AND job_id = ? ORDER BY created_at DESC LIMIT 1`
	case "mssql":
		query = `SELECT TOP 1 raw_json FROM reports WHERE agent_id = ? AND job_id = ? ORDER BY created_at DESC`
	default:
		query = `SELECT raw_json FROM reports WHERE agent_id = ? AND job_id = ? ORDER BY created_at DESC`
	}
	row := database.QueryRow(query, agentId, jobId)
	var raw string
	err = row.Scan(&raw)
	if err != nil {
		http.Error(w, "Aucun rapport trouvé", http.StatusNotFound)
		return
	}
	var report schema.DscReport
	if err := json.Unmarshal([]byte(raw), &report); err != nil {
		log.Printf("[API][REPORTS] Erreur parsing JSON: %v", err)
		http.Error(w, "Erreur parsing JSON", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}
