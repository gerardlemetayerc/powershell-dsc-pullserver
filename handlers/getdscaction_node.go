package handlers

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"go-dsc-pull/internal/scheduling"
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

	effectiveConfigs, scheduleType, err := resolveEffectiveConfigurations(database, agentId, time.Now().UTC())
	if err != nil {
		log.Printf("[GETDSCACTION-NODE] Erreur resolution configuration pour agent %s: %v", agentId, err)
	}
	if len(effectiveConfigs) == 0 {
		effectiveConfigs = []string{agentId}
	}
	if scheduleType != "none" {
		log.Printf("[GETDSCACTION-NODE] Configuration planifiee activee pour agent %s: %v (%s)", agentId, effectiveConfigs, scheduleType)
	}
	for _, cfg := range effectiveConfigs {
		if err := db.MarkConfigurationRequested(database, agentId, cfg); err != nil {
			log.Printf("[GETDSCACTION-NODE] Erreur memorisation configuration demandee pour agent=%s config=%s: %v", agentId, cfg, err)
		}
	}

	// Utiliser les schémas importés
	var cBody schema.ClientBody
	nodeStatus := "GetConfiguration"
	status := "GetConfiguration"
	if err := json.Unmarshal(body, &cBody); err == nil && len(cBody.ClientStatus) > 0 {
			checksumByConfig := make(map[string]string)
			for _, cs := range cBody.ClientStatus {
				if !strings.EqualFold(cs.ChecksumAlgorithm, "SHA-256") || cs.Checksum == "" {
					continue
				}
				checksumByConfig[strings.ToLower(strings.TrimSpace(cs.ConfigurationName))] = cs.Checksum
			}

			allComparableAndOk := true
			for _, cfg := range effectiveConfigs {
				row := database.QueryRow("SELECT mof_file FROM configuration_model WHERE LOWER(name) = LOWER(?)", cfg)
				var mofBytes []byte
				if err := row.Scan(&mofBytes); err == nil {
					hash := sha256SumHex(mofBytes)
					checksum, hasChecksum := checksumByConfig[strings.ToLower(cfg)]
					if hasChecksum {
						if !strings.EqualFold(hash, checksum) {
							allComparableAndOk = false
						}
					} else {
						allComparableAndOk = false
					}

					_, err = database.Exec("UPDATE configuration_model SET last_usage = CURRENT_TIMESTAMP WHERE LOWER(name) = LOWER(?)", cfg)
					if err != nil {
						log.Printf("[GETDSCACTION-NODE] Erreur mise a jour last_usage pour configName='%s': %v", cfg, err)
					}
				} else {
					allComparableAndOk = false
					log.Printf("[GETDSCACTION-NODE] MOF introuvable pour configName='%s': %v", cfg, err)
				}
			}

			if allComparableAndOk && len(effectiveConfigs) > 0 {
				status = "OK"
				nodeStatus = "OK"
				_, err = database.Exec("UPDATE agents SET last_communication = CURRENT_TIMESTAMP, state = 'OK' WHERE agent_id = ?", agentId)
				if err != nil {
					log.Printf("[GETDSCACTION-NODE] Erreur update last_communication state OK: %v", err)
				}
			}
	}

		details := make([]schema.DscActionDetail, 0, len(effectiveConfigs))
		for _, cfg := range effectiveConfigs {
			details = append(details, schema.DscActionDetail{
				ConfigurationName: cfg,
				Status:            status,
			})
		}

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

func resolveEffectiveConfigurations(database *sql.DB, agentId string, now time.Time) ([]string, string, error) {
	bindings, err := db.GetAgentConfigurationBindings(database, agentId)
	if err != nil {
		return nil, "none", err
	}
	if len(bindings) == 0 {
		return nil, "none", nil
	}

	primaries := make([]string, 0)
	bestScheduledName := ""
	bestScheduledType := "none"
	bestOccurrence := time.Time{}

	for _, binding := range bindings {
		scheduleType := strings.ToLower(strings.TrimSpace(binding.ScheduleType))
		if scheduleType == "none" && binding.Enabled {
			primaries = append(primaries, binding.ConfigurationName)
		}

		due, occurrence := scheduling.IsScheduleDue(binding, now)
		if due {
			if bestScheduledName == "" || occurrence.After(bestOccurrence) {
				bestScheduledName = binding.ConfigurationName
				bestScheduledType = scheduleType
				bestOccurrence = occurrence
			}
		}
	}

	if bestScheduledName != "" {
		err := db.MarkScheduledConfigurationApplied(database, agentId, bestScheduledName, bestScheduledType == "oneshot")
		if err != nil {
			log.Printf("[GETDSCACTION-NODE] Erreur maj scheduled_last_applied_at pour agent=%s config=%s: %v", agentId, bestScheduledName, err)
		}
		return []string{bestScheduledName}, bestScheduledType, nil
	}

	if len(primaries) > 0 {
		return primaries, "none", nil
	}

	return []string{bindings[0].ConfigurationName}, "none", nil
}
