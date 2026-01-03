package handlers

import (
	"encoding/json"
	"go-dsc-pull/internal/db"
	"net/http"
	"database/sql"
	"strings"
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

	// Récupère le filtre operationType si présent
	opType := r.URL.Query().Get("operationtype")
	var rows *sql.Rows
	var debugMsg string
	if opType != "" {
		// Filtre insensible à la casse
		opTypeLower := strings.ToLower(opType)
		debugMsg = "Filtre operationType: '" + opTypeLower + "'"
		rows, err = database.Query(`SELECT id, job_id, created_at, status FROM reports WHERE agent_id = ? AND LOWER(operation_type) = ? ORDER BY created_at DESC`, agentId, opTypeLower)
	} else {
		debugMsg = "Pas de filtre operationType"
		rows, err = database.Query(`SELECT id, job_id, created_at, status FROM reports WHERE agent_id = ? ORDER BY created_at DESC`, agentId)
	}
	if err != nil {
		http.Error(w, "DB query error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type ReportSummary struct {
		ID        int64  `json:"id"`
		JobId     string `json:"job_id"`
		CreatedAt string `json:"created_at"`
		Status    string `json:"status"`
	}
	reports := []ReportSummary{}
	count := 0
	for rows.Next() {
		var rep ReportSummary
		if err := rows.Scan(&rep.ID, &rep.JobId, &rep.CreatedAt, &rep.Status); err == nil {
			reports = append(reports, rep)
			count++
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Debug-OperationType", debugMsg+" | Results: "+string(rune(count)))
	_ = json.NewEncoder(w).Encode(reports)
}
