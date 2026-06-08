package db

import (
	"database/sql"
	"strings"
)

func EnsureInfraSchema(database *sql.DB, driver string) error {
	if driver == "mssql" || driver == "sqlserver" {
		return nil
	}

	stmts := []string{
		"ALTER TABLE dsc_infra_info ADD COLUMN latest_release TEXT",
		"ALTER TABLE dsc_infra_info ADD COLUMN latest_release_url TEXT",
		"ALTER TABLE dsc_infra_info ADD COLUMN update_available INTEGER DEFAULT 0",
		"ALTER TABLE dsc_infra_info ADD COLUMN release_check_ok INTEGER DEFAULT 0",
		"ALTER TABLE dsc_infra_info ADD COLUMN release_checked_at TIMESTAMP",
	}
	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			msg := strings.ToLower(err.Error())
			if strings.Contains(msg, "duplicate column name") {
				continue
			}
			return err
		}
	}
	return nil
}
