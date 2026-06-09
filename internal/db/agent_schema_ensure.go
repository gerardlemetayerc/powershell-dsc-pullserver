package db

import (
	"database/sql"
	"strings"
)

func EnsureAgentInternalDSCIdSchema(database *sql.DB, driver string) error {
	driverName := strings.ToLower(driver)
	if driverName == "mssql" || driverName == "sqlserver" {
		stmts := []string{
			"IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agents' AND COLUMN_NAME = 'internal_dsc_id') ALTER TABLE agents ADD internal_dsc_id NVARCHAR(128) NULL",
			"IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_agents_internal_dsc_id' AND object_id = OBJECT_ID('agents') AND has_filter = 1) DROP INDEX idx_agents_internal_dsc_id ON agents",
			"UPDATE agents SET internal_dsc_id = CONCAT('IDSC-', REPLACE(CONVERT(NVARCHAR(36), NEWID()), '-', '')) WHERE internal_dsc_id IS NULL",
			"IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_agents_internal_dsc_id' AND object_id = OBJECT_ID('agents')) CREATE UNIQUE INDEX idx_agents_internal_dsc_id ON agents(internal_dsc_id)",
		}
		for _, stmt := range stmts {
			if _, err := database.Exec(stmt); err != nil {
				return err
			}
		}
		return nil
	}

	stmts := []string{
		"ALTER TABLE agents ADD COLUMN internal_dsc_id TEXT",
		"UPDATE agents SET internal_dsc_id = 'IDSC-' || UPPER(REPLACE(agent_id, '-', '')) WHERE internal_dsc_id IS NULL",
		"CREATE UNIQUE INDEX IF NOT EXISTS idx_agents_internal_dsc_id ON agents(internal_dsc_id)",
	}

	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "duplicate column name") || strings.Contains(msg, "already exists") {
				continue
			}
			return err
		}
	}
	return nil
}
