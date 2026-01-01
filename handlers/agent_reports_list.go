package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal/db"
)

// AgentReportsListHandler retourne la liste des rapports d'un agent
func AgentReportsListHandler(w http.ResponseWriter, r *http.Request) {
	agentId := r.PathValue("id")
	if agentId == "" {
		http.Error(w, "AgentId manquant", http.StatusBadRequest)
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

	rows, err := database.Query(`SELECT id, job_id, created_at FROM reports WHERE agent_id = ? ORDER BY created_at DESC`, agentId)
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

type ReportSummary struct {
		ID        int64  `json:"id"`
		JobId     string `json:"job_id"`
		CreatedAt string `json:"created_at"`
	}
	reports := []ReportSummary{}
	for rows.Next() {
		var rep ReportSummary
		if err := rows.Scan(&rep.ID, &rep.JobId, &rep.CreatedAt); err == nil {
			reports = append(reports, rep)
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(reports)
}
