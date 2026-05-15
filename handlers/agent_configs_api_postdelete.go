package handlers

import (
	"encoding/json"
	"errors"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
	"log"
	"net/http"
	"strings"
	"time"
)

func parseScheduleTime(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, errors.New("invalid schedule time format")
}

// AgentConfigsAPIHandlerPostDelete gère POST (ajout) et DELETE (suppression) sur /api/v1/agents/{id}/configs
func AgentConfigsAPIHandlerPostDelete(w http.ResponseWriter, r *http.Request) {
	agentId := r.PathValue("id")
	if agentId == "" {
		http.Error(w, "AgentId manquant", http.StatusBadRequest)
		return
	}

	database, err := db.OpenDB(&global.AppConfig.Database)
	if err != nil {
		log.Printf("[API][DB] Erreur ouverture DB: %v", err)
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	switch r.Method {
	case http.MethodPost, http.MethodDelete:
		// TODO: Replace with new admin check middleware if needed
		if r.Method == http.MethodPost {
			var req struct {
				ConfigurationName string `json:"configuration_name"`
				ScheduleType      string `json:"schedule_type"`
				ScheduledAt       string `json:"scheduled_at"`
				RecurrenceMinutes *int   `json:"recurrence_minutes"`
				WindowMinutes     *int   `json:"window_minutes"`
				Enabled           *bool  `json:"enabled"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ConfigurationName == "" {
				http.Error(w, "Nom de configuration manquant ou invalide", http.StatusBadRequest)
				return
			}

			scheduleType := strings.ToLower(strings.TrimSpace(req.ScheduleType))
			if scheduleType == "" {
				scheduleType = "none"
			}
			if scheduleType != "none" && scheduleType != "oneshot" && scheduleType != "recurring" {
				http.Error(w, "schedule_type invalide (none|oneshot|recurring)", http.StatusBadRequest)
				return
			}

			windowMinutes := 30
			if req.WindowMinutes != nil {
				windowMinutes = *req.WindowMinutes
			}
			if windowMinutes <= 0 {
				http.Error(w, "window_minutes doit etre > 0", http.StatusBadRequest)
				return
			}

			enabled := true
			if req.Enabled != nil {
				enabled = *req.Enabled
			}

			var scheduledAt interface{} = nil
			var recurrenceMinutes interface{} = nil
			if scheduleType == "oneshot" || scheduleType == "recurring" {
				if strings.TrimSpace(req.ScheduledAt) == "" {
					http.Error(w, "scheduled_at requis pour une configuration planifiee", http.StatusBadRequest)
					return
				}
				t, err := parseScheduleTime(strings.TrimSpace(req.ScheduledAt))
				if err != nil {
					http.Error(w, "Format scheduled_at invalide", http.StatusBadRequest)
					return
				}
				// Pass a native time value so SQL Server receives a datetime parameter, not NVARCHAR.
				scheduledAt = t.UTC()

				if scheduleType == "recurring" {
					if req.RecurrenceMinutes == nil || *req.RecurrenceMinutes <= 0 {
						http.Error(w, "recurrence_minutes doit etre > 0 pour recurring", http.StatusBadRequest)
						return
					}
					recurrenceMinutes = *req.RecurrenceMinutes
				}
			}

			driver := global.AppConfig.Database.Driver
			if driver == "sqlite" {
				_, err := database.Exec(`
					   INSERT OR REPLACE INTO agent_configurations (
						   agent_id,
						   configuration_name,
						   schedule_type,
						   scheduled_at,
						   recurrence_minutes,
						   window_minutes,
						   enabled
					   )
					   VALUES (?, ?, ?, ?, ?, ?, ?)
				   `, agentId, req.ConfigurationName, scheduleType, scheduledAt, recurrenceMinutes, windowMinutes, enabled)
				if err != nil {
					log.Printf("[API][CONFIG][POST] Erreur insertion sqlite agent=%s config=%s: %v", agentId, req.ConfigurationName, err)
					http.Error(w, "Erreur insertion config", http.StatusInternalServerError)
					return
				}
			} else {
				res, err := database.Exec(`
					UPDATE agent_configurations
					SET schedule_type = ?,
						scheduled_at = ?,
						recurrence_minutes = ?,
						window_minutes = ?,
						enabled = ?
					WHERE agent_id = ?
					  AND configuration_name = ?
				`, scheduleType, scheduledAt, recurrenceMinutes, windowMinutes, enabled, agentId, req.ConfigurationName)
				if err != nil {
					log.Printf("[API][CONFIG][POST] Erreur update mssql agent=%s config=%s: %v", agentId, req.ConfigurationName, err)
					http.Error(w, "Erreur insertion config", http.StatusInternalServerError)
					return
				}

				rowsAffected, err := res.RowsAffected()
				if err != nil {
					log.Printf("[API][CONFIG][POST] Erreur RowsAffected mssql agent=%s config=%s: %v", agentId, req.ConfigurationName, err)
					http.Error(w, "Erreur insertion config", http.StatusInternalServerError)
					return
				}

				if rowsAffected == 0 {
					_, err = database.Exec(`
						INSERT INTO agent_configurations (
							agent_id,
							configuration_name,
							schedule_type,
							scheduled_at,
							recurrence_minutes,
							window_minutes,
							enabled
						) VALUES (?, ?, ?, ?, ?, ?, ?)
					`, agentId, req.ConfigurationName, scheduleType, scheduledAt, recurrenceMinutes, windowMinutes, enabled)
					if err != nil {
						log.Printf("[API][CONFIG][POST] Erreur insert mssql agent=%s config=%s: %v", agentId, req.ConfigurationName, err)
						http.Error(w, "Erreur insertion config", http.StatusInternalServerError)
						return
					}
				}
			}
			w.WriteHeader(http.StatusCreated)
		} else {
			var req struct {
				ConfigurationName string `json:"configuration_name"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ConfigurationName == "" {
				http.Error(w, "Nom de configuration manquant ou invalide", http.StatusBadRequest)
				return
			}
			_, err := database.Exec(`DELETE FROM agent_configurations WHERE agent_id = ? AND configuration_name = ?`, agentId, req.ConfigurationName)
			if err != nil {
				http.Error(w, "Erreur suppression config", http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNoContent)
		}
	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}
