package handlers

import (
	"io"
	"log"
	"strings"
	"net/http"
	"encoding/json"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/utils"
	"path/filepath"
)

// SendReportHandler gère POST /PSDSCPullServer.svc/Nodes(AgentId='...')/SendReport
func SendReportHandler(w http.ResponseWriter, r *http.Request) {
	reportBody, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[SENDREPORT] Erreur lecture body: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	rawId := r.PathValue("node")
	agentId := utils.ExtractAgentId(rawId)
	log.Printf("[SENDREPORT] AgentId=%s (raw=%s), ReportSize=%d", agentId, rawId, len(reportBody))

	// Désérialiser le rapport
	var report schema.DscReport
	if err := json.Unmarshal(reportBody, &report); err != nil {
		log.Printf("[SENDREPORT] Erreur parsing JSON rapport: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Préparer les champs JSON pour la base
	errorsJson, _ := json.Marshal(report.Errors)
	statusDataJson, _ := json.Marshal(report.StatusData)
	additionalDataJson, _ := json.Marshal(report.AdditionalData)

	// Insérer en base
	   exeDir, err := utils.ExePath()
	   var dbCfg *schema.DBConfig
	   if err == nil {
		   configPath := filepath.Join(filepath.Dir(exeDir), "config.json")
		   dbCfg, err = db.LoadDBConfig(configPath)
	   }
	   if err == nil {
		   database, err := db.OpenDB(dbCfg)
		if err == nil {
			   // Vérifie si un rapport existe déjà pour ce job_id
			   var count int
			   err = database.QueryRow("SELECT COUNT(*) FROM reports WHERE job_id = ?", report.JobId).Scan(&count)
			   if err != nil {
				   log.Printf("[SENDREPORT] Erreur SELECT COUNT sur reports: %v", err)
			   }
			   if count > 0 {
				   log.Printf("[SENDREPORT] Update rapport en base: agent_id=%s, job_id=%s, operation_type=%s", agentId, report.JobId, report.OperationType)
				   _, err := database.Exec(`UPDATE reports SET 
					   agent_id=?, report_format_version=?, operation_type=?, refresh_mode=?, status=?, start_time=?, end_time=?, reboot_requested=?, errors=?, status_data=?, additional_data=?, raw_json=?
					   WHERE job_id=?`,
					   agentId,
					   report.ReportFormatVersion,
					   report.OperationType,
					   report.RefreshMode,
					   report.Status,
					   report.StartTime,
					   report.EndTime,
					   report.RebootRequested,
					   string(errorsJson),
					   string(statusDataJson),
					   string(additionalDataJson),
					   string(reportBody),
					   report.JobId,
				   )
				   if err != nil {
					   log.Printf("[SENDREPORT] Erreur update rapport en base: %v", err)
				   }
			   } else {
				   log.Printf("[SENDREPORT] Insertion rapport en base: agent_id=%s, job_id=%s, operation_type=%s", agentId, report.JobId, report.OperationType)
				   _, err := database.Exec(`INSERT INTO reports (
					   agent_id, job_id, report_format_version, operation_type, refresh_mode, status, start_time, end_time, reboot_requested, errors, status_data, additional_data, raw_json
				   ) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
					   agentId,
					   report.JobId,
					   report.ReportFormatVersion,
					   report.OperationType,
					   report.RefreshMode,
					   report.Status,
					   report.StartTime,
					   report.EndTime,
					   report.RebootRequested,
					   string(errorsJson),
					   string(statusDataJson),
					   string(additionalDataJson),
					   string(reportBody),
				   )
				   if err != nil {
					   log.Printf("[SENDREPORT] Erreur insertion rapport en base: %v", err)
				   }
			   }
			   // Met à jour last_communication et has_error_last_report uniquement pour les rapports Initial
			   hasError := 0
			   if report.OperationType == "Initial" {
					log.Printf("[SENDREPORT] Initial report received, updating last_communication and has_error_last_report")
				   if strings.ToLower(report.Status) == "failure" {
					   hasError = 1
				   }
				   _, err = database.Exec("UPDATE agents SET last_communication = CURRENT_TIMESTAMP, has_error_last_report = ?, state = ? WHERE agent_id = ?", hasError, report.Status, agentId)
				   if err != nil {
					   log.Printf("[SENDREPORT] Erreur update last_communication/has_error_last_report: %v", err)
				   }
			   } else {
				   // Pour les autres types de rapport, on met juste à jour last_communication
				   _, err = database.Exec("UPDATE agents SET last_communication = CURRENT_TIMESTAMP WHERE agent_id = ?", agentId)
				   if err != nil {
					   log.Printf("[SENDREPORT] Erreur update last_communication: %v", err)
				   }
			   }
			database.Close()
		}
	}

	
	w.Header().Set("ProtocolVersion", "2.0")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
