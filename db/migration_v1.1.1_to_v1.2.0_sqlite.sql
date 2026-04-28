-- Migration script from v1.1.1 to v1.2.0 (SQLite)
-- Adds scheduling support for agent configurations.

ALTER TABLE agent_configurations ADD COLUMN schedule_type TEXT NOT NULL DEFAULT 'none';
ALTER TABLE agent_configurations ADD COLUMN scheduled_at TIMESTAMP NULL;
ALTER TABLE agent_configurations ADD COLUMN recurrence_minutes INTEGER NULL;
ALTER TABLE agent_configurations ADD COLUMN window_minutes INTEGER NOT NULL DEFAULT 30;
ALTER TABLE agent_configurations ADD COLUMN scheduled_last_applied_at TIMESTAMP NULL;
ALTER TABLE agent_configurations ADD COLUMN last_requested_at TIMESTAMP NULL;
ALTER TABLE agent_configurations ADD COLUMN last_execution_status TEXT NULL;
ALTER TABLE agent_configurations ADD COLUMN last_execution_at TIMESTAMP NULL;
ALTER TABLE agent_configurations ADD COLUMN enabled BOOLEAN NOT NULL DEFAULT 1;

UPDATE dsc_infra_info SET db_version = '1.2.0', updated_at = CURRENT_TIMESTAMP WHERE id = 1;
