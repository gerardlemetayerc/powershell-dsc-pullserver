package db

import (
	"database/sql"
)

// GetAgentConfigurations retourne la liste des configurations associées à un agent
func GetAgentConfigurations(db *sql.DB, agentId string) ([]string, error) {
	rows, err := db.Query(`SELECT configuration_name FROM agent_configurations WHERE agent_id = ?`, agentId)
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
