package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/schema"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// GetDscActionNodeHandlerWithId gère POST /PSDSCPullServer.svc/{id}/GetDscAction avec agentId déjà extrait
func GetDscActionNodeHandlerWithId(w http.ResponseWriter, r *http.Request, agentId string) {
	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		log.Printf("[GETDSCACTION-NODE] Erreur ouverture base: %v", err)
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Log du body et des headers reçus pour debug
	body, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// DEBUG : afficher le body dans la réponse HTTP (en plus du log)
	w.Header().Set("X-Debug-Client-Body", string(body))

	effectiveConfig, scheduleType, err := resolveEffectiveConfiguration(database, agentId, time.Now().UTC())
	if err != nil {
		log.Printf("[GETDSCACTION-NODE] Erreur resolution configuration pour agent %s: %v", agentId, err)
	}
	if effectiveConfig == "" {
		effectiveConfig = agentId
	}
	if scheduleType != "none" {
		log.Printf("[GETDSCACTION-NODE] Configuration planifiee activee pour agent %s: %s (%s)", agentId, effectiveConfig, scheduleType)
	}

	// Utiliser les schémas importés
	var cBody schema.ClientBody
	nodeStatus := "GetConfiguration"
	status := "GetConfiguration"
	if err := json.Unmarshal(body, &cBody); err == nil && len(cBody.ClientStatus) > 0 {
		row := database.QueryRow("SELECT mof_file FROM configuration_model WHERE LOWER(name) = LOWER(?)", effectiveConfig)
		var mofBytes []byte
		if err := row.Scan(&mofBytes); err == nil {
			hash := sha256SumHex(mofBytes)
			allOk := true
			hasComparableStatus := false
			for _, cs := range cBody.ClientStatus {
				if strings.EqualFold(cs.ChecksumAlgorithm, "SHA-256") && cs.Checksum != "" {
					hasComparableStatus = true
					if !strings.EqualFold(hash, cs.Checksum) {
						allOk = false
						break
					}
				} else {
					allOk = false
					break
				}
			}

			if hasComparableStatus && allOk {
				status = "OK"
				nodeStatus = "OK"
				_, err = database.Exec("UPDATE agents SET last_communication = CURRENT_TIMESTAMP, state = 'OK' WHERE agent_id = ?", agentId)
				if err != nil {
					log.Printf("[GETDSCACTION-NODE] Erreur update last_communication state OK: %v", err)
				}
			}

			_, err = database.Exec("UPDATE configuration_model SET last_usage = CURRENT_TIMESTAMP WHERE LOWER(name) = LOWER(?)", effectiveConfig)
			if err != nil {
				log.Printf("[GETDSCACTION-NODE] Erreur mise a jour last_usage pour configName='%s': %v", effectiveConfig, err)
			}
		} else {
			log.Printf("[GETDSCACTION-NODE] MOF introuvable pour configName='%s': %v", effectiveConfig, err)
		}
	}

	details := []schema.DscActionDetail{{
		ConfigurationName: effectiveConfig,
		Status:            status,
	}}

	resp := schema.DscActionResponse{
		NodeStatus: nodeStatus,
		Details:    details,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ProtocolVersion", "2.0")
	json.NewEncoder(w).Encode(resp)
}

// sha256SumHex calcule le hash SHA-256 d'un tableau d'octets et retourne l'hexadécimal en majuscules
func sha256SumHex(data []byte) string {
	h := sha256.Sum256(data)
	return strings.ToUpper(fmt.Sprintf("%X", h[:]))
}

func parseDBTime(value string) (time.Time, bool) {
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func isScheduleDue(binding schema.AgentConfigurationBinding, now time.Time) (bool, time.Time) {
	scheduleType := strings.ToLower(strings.TrimSpace(binding.ScheduleType))
	if scheduleType != "oneshot" && scheduleType != "recurring" {
		return false, time.Time{}
	}
	if !binding.Enabled || binding.ScheduledAt == nil || *binding.ScheduledAt == "" {
		return false, time.Time{}
	}

	startAt, ok := parseDBTime(*binding.ScheduledAt)
	if !ok {
		return false, time.Time{}
	}

	windowMinutes := binding.WindowMinutes
	if windowMinutes <= 0 {
		windowMinutes = 30
	}
	window := time.Duration(windowMinutes) * time.Minute
	if now.Before(startAt) {
		return false, time.Time{}
	}

	if scheduleType == "oneshot" {
		if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
			return false, time.Time{}
		}
		if now.After(startAt.Add(window)) {
			return false, time.Time{}
		}
		return true, startAt
	}

	if binding.RecurrenceMinutes == nil || *binding.RecurrenceMinutes <= 0 {
		return false, time.Time{}
	}

	interval := time.Duration(*binding.RecurrenceMinutes) * time.Minute
	elapsed := now.Sub(startAt)
	occurrenceCount := int64(elapsed / interval)
	occurrenceStart := startAt.Add(time.Duration(occurrenceCount) * interval)
	if now.After(occurrenceStart.Add(window)) {
		return false, time.Time{}
	}

	if binding.ScheduledLastAppliedAt != nil && *binding.ScheduledLastAppliedAt != "" {
		if lastApplied, ok := parseDBTime(*binding.ScheduledLastAppliedAt); ok {
			if !lastApplied.Before(occurrenceStart) {
				return false, time.Time{}
			}
		}
	}

	return true, occurrenceStart
}

func resolveEffectiveConfiguration(database *sql.DB, agentId string, now time.Time) (string, string, error) {
	bindings, err := db.GetAgentConfigurationBindings(database, agentId)
	if err != nil {
		return "", "none", err
	}
	if len(bindings) == 0 {
		return "", "none", nil
	}

	primary := ""
	bestScheduledName := ""
	bestScheduledType := "none"
	bestOccurrence := time.Time{}

	for _, binding := range bindings {
		scheduleType := strings.ToLower(strings.TrimSpace(binding.ScheduleType))
		if scheduleType == "none" && binding.Enabled && primary == "" {
			primary = binding.ConfigurationName
		}

		due, occurrence := isScheduleDue(binding, now)
		if due {
			if bestScheduledName == "" || occurrence.After(bestOccurrence) {
				bestScheduledName = binding.ConfigurationName
				bestScheduledType = scheduleType
				bestOccurrence = occurrence
			}
		}
	}

	if primary == "" {
		primary = bindings[0].ConfigurationName
	}

	if bestScheduledName != "" {
		err := db.MarkScheduledConfigurationApplied(database, agentId, bestScheduledName, bestScheduledType == "oneshot")
		if err != nil {
			log.Printf("[GETDSCACTION-NODE] Erreur maj scheduled_last_applied_at pour agent=%s config=%s: %v", agentId, bestScheduledName, err)
		}
		return bestScheduledName, bestScheduledType, nil
	}

	return primary, "none", nil
}
