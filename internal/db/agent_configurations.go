package db

import (
	"database/sql"
	"go-dsc-pull/internal/schema"
	"strings"
)

func execWithUTCNowFallback(db *sql.DB, queryTemplate string, args ...interface{}) error {
	query := strings.ReplaceAll(queryTemplate, "{{NOW_UTC}}", "SYSUTCDATETIME()")
	_, err := db.Exec(query, args...)
	if err == nil {
		return nil
	}

	query = strings.ReplaceAll(queryTemplate, "{{NOW_UTC}}", "CURRENT_TIMESTAMP")
	_, err = db.Exec(query, args...)
	return err
}

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
			last_requested_at,
			last_execution_status,
			last_execution_at,
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
		var lastRequestedAt, lastExecutionStatus, lastExecutionAt sql.NullString
		var recurrenceMinutes sql.NullInt64
		if err := rows.Scan(
			&b.AgentID,
			&b.ConfigurationName,
			&b.ScheduleType,
			&scheduledAt,
			&recurrenceMinutes,
			&b.WindowMinutes,
			&scheduledLastAppliedAt,
			&lastRequestedAt,
			&lastExecutionStatus,
			&lastExecutionAt,
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
		if lastRequestedAt.Valid {
			v := lastRequestedAt.String
			b.LastRequestedAt = &v
		}
		if lastExecutionStatus.Valid {
			v := lastExecutionStatus.String
			b.LastExecutionStatus = &v
		}
		if lastExecutionAt.Valid {
			v := lastExecutionAt.String
			b.LastExecutionAt = &v
		}
		bindings = append(bindings, b)
	}

	return bindings, nil
}

// MarkScheduledConfigurationApplied met a jour l'etat d'application d'une configuration planifiee.
func MarkScheduledConfigurationApplied(db *sql.DB, agentId, configurationName string, disable bool) error {
	if disable {
		err := execWithUTCNowFallback(db, `
			UPDATE agent_configurations
			SET scheduled_last_applied_at = {{NOW_UTC}},
				enabled = 0
			WHERE agent_id = ?
			  AND configuration_name = ?
			  AND schedule_type = 'oneshot'
		`, agentId, configurationName)
		return err
	}

	err := execWithUTCNowFallback(db, `
		UPDATE agent_configurations
		SET scheduled_last_applied_at = {{NOW_UTC}}
		WHERE agent_id = ?
		  AND configuration_name = ?
		  AND schedule_type IN ('oneshot', 'recurring')
	`, agentId, configurationName)
	return err
}

// MarkConfigurationRequested memorise la derniere configuration demandee a un noeud.
func MarkConfigurationRequested(db *sql.DB, agentId, configurationName string) error {
	err := execWithUTCNowFallback(db, `
		UPDATE agent_configurations
		SET last_requested_at = {{NOW_UTC}}
		WHERE agent_id = ?
		  AND configuration_name = ?
	`, agentId, configurationName)
	return err
}

// UpdateLastConfigurationExecutionStatus rattache le statut du dernier rapport a la configuration servie la plus recente.
func UpdateLastConfigurationExecutionStatus(db *sql.DB, agentId, status string) error {
	err := execWithUTCNowFallback(db, `
		UPDATE agent_configurations
		SET last_execution_status = ?,
			last_execution_at = {{NOW_UTC}}
		WHERE agent_id = ?
		  AND configuration_name = (
				SELECT configuration_name
				FROM agent_configurations
				WHERE agent_id = ?
				  AND last_requested_at IS NOT NULL
				ORDER BY last_requested_at DESC
				LIMIT 1
		  )
	`, status, agentId, agentId)
	if err == nil {
		return nil
	}

	err = execWithUTCNowFallback(db, `
		UPDATE agent_configurations
		SET last_execution_status = ?,
			last_execution_at = {{NOW_UTC}}
		WHERE agent_id = ?
		  AND configuration_name = (
				SELECT TOP 1 configuration_name
				FROM agent_configurations
				WHERE agent_id = ?
				  AND last_requested_at IS NOT NULL
				ORDER BY last_requested_at DESC
		  )
	`, status, agentId, agentId)
	return err
}

// UpdateConfigurationExecutionStatusByName met a jour l'execution pour une configuration precise.
func UpdateConfigurationExecutionStatusByName(db *sql.DB, agentId, configurationName, status string) error {
	err := execWithUTCNowFallback(db, `
		UPDATE agent_configurations
		SET last_execution_status = ?,
			last_execution_at = {{NOW_UTC}}
		WHERE agent_id = ?
		  AND LOWER(configuration_name) = LOWER(?)
	`, status, agentId, configurationName)
	if err == nil {
		return nil
	}

	_, err = db.Exec(`
		UPDATE agent_configurations
		SET last_execution_status = ?,
			last_execution_at = CURRENT_TIMESTAMP
		WHERE agent_id = ?
		  AND LOWER(configuration_name) = LOWER(?)
	`, status, agentId, configurationName)
	return err
}
