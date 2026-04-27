package db

import (
	"database/sql"
	"go-dsc-pull/internal/schema"
)

// GetAgentConfigurations retourne la liste des configurations associées à un agent
func GetAgentConfigurations(db *sql.DB, agentId string) ([]string, error) {
	rows, err := db.Query(`SELECT configuration_name FROM agent_configurations WHERE agent_id = ? AND enabled = 1`, agentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var configs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			configs = append(configs, name)
		}
	}
	return configs, nil
}

// GetAgentConfigurationBindings retourne toutes les associations de configuration d'un agent.
func GetAgentConfigurationBindings(db *sql.DB, agentId string) ([]schema.AgentConfigurationBinding, error) {
	rows, err := db.Query(`
		SELECT
			agent_id,
			configuration_name,
			schedule_type,
			scheduled_at,
			recurrence_minutes,
			window_minutes,
			scheduled_last_applied_at,
			enabled
		FROM agent_configurations
		WHERE agent_id = ?
		ORDER BY CASE WHEN schedule_type = 'none' THEN 0 WHEN schedule_type = 'oneshot' THEN 1 ELSE 2 END, configuration_name
	`, agentId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bindings := make([]schema.AgentConfigurationBinding, 0)
	for rows.Next() {
		var b schema.AgentConfigurationBinding
		var scheduledAt, scheduledLastAppliedAt sql.NullString
		var recurrenceMinutes sql.NullInt64
		if err := rows.Scan(
			&b.AgentID,
			&b.ConfigurationName,
			&b.ScheduleType,
			&scheduledAt,
			&recurrenceMinutes,
			&b.WindowMinutes,
			&scheduledLastAppliedAt,
			&b.Enabled,
		); err != nil {
			continue
		}
		if scheduledAt.Valid {
			v := scheduledAt.String
			b.ScheduledAt = &v
		}
		if recurrenceMinutes.Valid {
			v := int(recurrenceMinutes.Int64)
			b.RecurrenceMinutes = &v
		}
		if scheduledLastAppliedAt.Valid {
			v := scheduledLastAppliedAt.String
			b.ScheduledLastAppliedAt = &v
		}
		bindings = append(bindings, b)
	}

	return bindings, nil
}

// MarkScheduledConfigurationApplied met a jour l'etat d'application d'une configuration planifiee.
func MarkScheduledConfigurationApplied(db *sql.DB, agentId, configurationName string, disable bool) error {
	if disable {
		_, err := db.Exec(`
			UPDATE agent_configurations
			SET scheduled_last_applied_at = CURRENT_TIMESTAMP,
				enabled = 0
			WHERE agent_id = ?
			  AND configuration_name = ?
			  AND schedule_type = 'oneshot'
		`, agentId, configurationName)
		return err
	}

	_, err := db.Exec(`
		UPDATE agent_configurations
		SET scheduled_last_applied_at = CURRENT_TIMESTAMP
		WHERE agent_id = ?
		  AND configuration_name = ?
		  AND schedule_type IN ('oneshot', 'recurring')
	`, agentId, configurationName)
	return err
}
