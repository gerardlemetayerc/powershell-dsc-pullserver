package handlers

import (
	"fmt"
	"io"
	"log"
	"strings"
	"net/http"
	"encoding/json"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/utils"
)

func extractConfigurationNameFromMap(entry map[string]interface{}) string {
	for k, v := range entry {
		if strings.EqualFold(k, "ConfigurationName") {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" && s != "<nil>" {
				return s
			}
		}

		switch vv := v.(type) {
		case map[string]interface{}:
			if found := extractConfigurationNameFromMap(vv); found != "" {
				return found
			}
		case []interface{}:
			for _, item := range vv {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					if found := extractConfigurationNameFromMap(nestedMap); found != "" {
						return found
					}
				}
			}
		case string:
			trimmed := strings.TrimSpace(vv)
			if trimmed == "" || !(strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[")) {
				continue
			}
			var nested interface{}
			if err := json.Unmarshal([]byte(trimmed), &nested); err != nil {
				continue
			}
			switch n := nested.(type) {
			case map[string]interface{}:
				if found := extractConfigurationNameFromMap(n); found != "" {
					return found
				}
			case []interface{}:
				for _, item := range n {
					if nestedMap, ok := item.(map[string]interface{}); ok {
						if found := extractConfigurationNameFromMap(nestedMap); found != "" {
							return found
						}
					}
				}
			}
		}
	}

	return ""
}

func extractConfigurationNameFromStatusData(statusData []string) string {
	for _, statusEntry := range statusData {
		var entryMap map[string]interface{}
		if err := json.Unmarshal([]byte(statusEntry), &entryMap); err != nil {
			continue
		}
		if found := extractConfigurationNameFromMap(entryMap); found != "" {
			return found
		}
	}
	return ""
}

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

	// Contrôle de la présence de MetaData dans StatusData
	mofApplied := 0
	for _, statusEntry := range report.StatusData {
		var entryMap map[string]interface{}
		if err := json.Unmarshal([]byte(statusEntry), &entryMap); err == nil {
			if _, ok := entryMap["MetaData"]; ok {
				mofApplied = 1
				break
			}
		}
	}

	// Insérer en base
	if err == nil {
		database, err := db.OpenDB(&global.AppConfig.Database)
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
					agent_id=?, report_format_version=?, operation_type=?, refresh_mode=?, status=?, start_time=?, end_time=?, reboot_requested=?, errors=?, status_data=?, additional_data=?, mof_applied=?, raw_json=?
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
					mofApplied,
					string(reportBody),
					report.JobId,
				)
				if err != nil {
					log.Printf("[SENDREPORT] Erreur update rapport en base: %v", err)
				}
			} else {
				log.Printf("[SENDREPORT] Insertion rapport en base: agent_id=%s, job_id=%s, operation_type=%s", agentId, report.JobId, report.OperationType)
				_, err := database.Exec(`INSERT INTO reports (
					agent_id, job_id, report_format_version, operation_type, refresh_mode, status, start_time, end_time, reboot_requested, errors, status_data, additional_data, mof_applied, raw_json
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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
					mofApplied,
					string(reportBody),
				)
				if err != nil {
					log.Printf("[SENDREPORT] Erreur insertion rapport en base: %v", err)
				}
			}
			reportedConfigurationName := extractConfigurationNameFromStatusData(report.StatusData)
			if reportedConfigurationName != "" {
				err = db.UpdateConfigurationExecutionStatusByName(database, agentId, reportedConfigurationName, report.Status)
				if err != nil {
					log.Printf("[SENDREPORT] Erreur update last_execution_status config=%s: %v", reportedConfigurationName, err)
				}
			} else if report.OperationType == "Initial" || mofApplied == 1 {
				err = db.UpdateLastConfigurationExecutionStatus(database, agentId, report.Status)
				if err != nil {
					log.Printf("[SENDREPORT] Erreur update last_execution_status (fallback): %v", err)
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
