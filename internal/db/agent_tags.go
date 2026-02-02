package db

import (
	"database/sql"
)


// SetAgentTag ajoute une valeur à un tag clé pour un agent (plusieurs valeurs possibles)
// Ajoutez driverName comme argument pour choisir la syntaxe
func SetAgentTag(db *sql.DB, driverName, agentId, key, value string) error {
       var err error
       if driverName == "sqlite" {
	       _, err = db.Exec(`INSERT OR IGNORE INTO agent_tags (agent_id, tag_key, tag_value) VALUES (?, ?, ?)`, agentId, key, value)
       } else {
	       _, err = db.Exec(`IF NOT EXISTS (SELECT 1 FROM agent_tags WHERE agent_id = ? AND tag_key = ? AND tag_value = ?) INSERT INTO agent_tags (agent_id, tag_key, tag_value) VALUES (?, ?, ?)`, agentId, key, value, agentId, key, value)
       }
       return err
}

// DeleteAgentTag supprime une valeur précise d'un tag clé pour un agent
func DeleteAgentTag(db *sql.DB, agentId, key, value string) error {
       _, err := db.Exec(`DELETE FROM agent_tags WHERE agent_id = ? AND tag_key = ? AND tag_value = ?`, agentId, key, value)
       return err
}

// GetAgentTags retourne tous les tags clé/valeurs pour un agent (clé -> tableau de valeurs)
func GetAgentTags(db *sql.DB, agentId string) (map[string][]string, error) {
       rows, err := db.Query(`SELECT tag_key, tag_value FROM agent_tags WHERE agent_id = ?`, agentId)
       if err != nil {
	       return nil, err
       }
       defer rows.Close()
       tags := make(map[string][]string)
       for rows.Next() {
	       var k, v string
	       if err := rows.Scan(&k, &v); err == nil {
		       tags[k] = append(tags[k], v)
	       }
       }
       return tags, nil
}

// ListAgentsByTag retourne les agent_id ayant un tag clé/valeur donné
func ListAgentsByTag(db *sql.DB, key, value string) ([]string, error) {
	rows, err := db.Query(`SELECT agent_id FROM agent_tags WHERE tag_key = ? AND tag_value = ?`, key, value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}
