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

func extractFirstErrorDetails(errors []string) (string, string) {
	for _, errorEntry := range errors {
		var errorMap map[string]interface{}
		if err := json.Unmarshal([]byte(errorEntry), &errorMap); err != nil {
			continue
		}
		errorCode := strings.TrimSpace(fmt.Sprint(errorMap["ErrorCode"]))
		errorMessage := strings.TrimSpace(fmt.Sprint(errorMap["ErrorMessage"]))
		if errorCode != "" && errorCode != "<nil>" {
			return errorCode, errorMessage
		}
		if errorMessage != "" && errorMessage != "<nil>" {
			return "", errorMessage
		}
	}
	return "", ""
}

func extractConfigurationNamesFromMap(entry map[string]interface{}, seen map[string]bool, names *[]string) {
	for k, v := range entry {
		if strings.EqualFold(k, "ConfigurationName") {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s != "" && s != "<nil>" {
				key := strings.ToLower(s)
				if !seen[key] {
					seen[key] = true
					*names = append(*names, s)
				}
			}
		}

		switch vv := v.(type) {
		case map[string]interface{}:
			extractConfigurationNamesFromMap(vv, seen, names)
		case []interface{}:
			for _, item := range vv {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					extractConfigurationNamesFromMap(nestedMap, seen, names)
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
				extractConfigurationNamesFromMap(n, seen, names)
			case []interface{}:
				for _, item := range n {
					if nestedMap, ok := item.(map[string]interface{}); ok {
						extractConfigurationNamesFromMap(nestedMap, seen, names)
					}
				}
			}
		}
	}
}

func extractConfigurationNamesFromStatusData(statusData []string) []string {
	seen := make(map[string]bool)
	names := make([]string, 0)
	for _, statusEntry := range statusData {
		var entryMap map[string]interface{}
		if err := json.Unmarshal([]byte(statusEntry), &entryMap); err != nil {
			continue
		}
		extractConfigurationNamesFromMap(entryMap, seen, &names)
	}
	return names
}

func extractConfigurationNameFromAdditionalData(additionalData []schema.DscKeyValue) string {
	for _, item := range additionalData {
		if strings.EqualFold(strings.TrimSpace(item.Key), "ConfigurationName") {
			value := strings.TrimSpace(item.Value)
			if value != "" && value != "<nil>" {
				return value
			}
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

	report.OperationType = strings.TrimSpace(report.OperationType)
	report.Status = strings.TrimSpace(report.Status)

	if report.OperationType == "" {
		report.OperationType = "Unknown"
	}

	if report.Status == "" {
		errorCode, errorMessage := extractFirstErrorDetails(report.Errors)
		report.Status = "Unknown"
		if errorCode != "" && errorCode != "0" {
			report.Status = "Failure"
		}
		log.Printf("[SENDREPORT] Report incomplet recu: job_id=%s operation_type=%s status=%s error_code=%s error_message=%s", report.JobId, report.OperationType, report.Status, errorCode, errorMessage)
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
			reportedConfigurationNames := extractConfigurationNamesFromStatusData(report.StatusData)
			if len(reportedConfigurationNames) == 0 {
				if cfgFromAdditionalData := extractConfigurationNameFromAdditionalData(report.AdditionalData); cfgFromAdditionalData != "" {
					reportedConfigurationNames = append(reportedConfigurationNames, cfgFromAdditionalData)
				}
			}
			if len(reportedConfigurationNames) > 0 {
				for _, cfgName := range reportedConfigurationNames {
					err = db.UpdateConfigurationExecutionStatusByName(database, agentId, cfgName, report.Status)
					if err != nil {
						log.Printf("[SENDREPORT] Erreur update last_execution_status config=%s: %v", cfgName, err)
					}
				}
			} else if report.OperationType == "Initial" || mofApplied == 1 {
				singleConfigName, uniqueConfig, lookupErr := db.GetSingleEnabledConfigurationName(database, agentId)
				if lookupErr != nil {
					log.Printf("[SENDREPORT] Erreur lecture configuration active unique: %v", lookupErr)
				} else if uniqueConfig {
					err = db.UpdateConfigurationExecutionStatusByName(database, agentId, singleConfigName, report.Status)
					if err != nil {
						log.Printf("[SENDREPORT] Erreur update last_execution_status (fallback unique) config=%s: %v", singleConfigName, err)
					}
				} else {
					log.Printf("[SENDREPORT] Fallback statut ignore: configuration cible ambigue (agent=%s, job_id=%s)", agentId, report.JobId)
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
		} else {
			log.Printf("[SENDREPORT] Erreur ouverture DB: %v", err)
		}
	}

	
	w.Header().Set("ProtocolVersion", "2.0")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{}"))
}
