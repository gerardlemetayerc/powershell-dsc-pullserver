-- Migration script from v1.1.1 to v1.2.0 (MSSQL)
-- Adds scheduling support for agent configurations.

IF EXISTS (SELECT * FROM sysobjects WHERE name='agent_configurations' AND xtype='U')
BEGIN
    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'schedule_type')
        ALTER TABLE agent_configurations ADD schedule_type NVARCHAR(16) NOT NULL CONSTRAINT DF_agent_configurations_schedule_type DEFAULT 'none';

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'scheduled_at')
        ALTER TABLE agent_configurations ADD scheduled_at DATETIME NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'recurrence_minutes')
        ALTER TABLE agent_configurations ADD recurrence_minutes INT NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'window_minutes')
        ALTER TABLE agent_configurations ADD window_minutes INT NOT NULL CONSTRAINT DF_agent_configurations_window_minutes DEFAULT 30;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'scheduled_last_applied_at')
        ALTER TABLE agent_configurations ADD scheduled_last_applied_at DATETIME NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'last_requested_at')
        ALTER TABLE agent_configurations ADD last_requested_at DATETIME NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'last_execution_status')
        ALTER TABLE agent_configurations ADD last_execution_status NVARCHAR(32) NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'last_execution_at')
        ALTER TABLE agent_configurations ADD last_execution_at DATETIME NULL;

    IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agent_configurations' AND COLUMN_NAME = 'enabled')
        ALTER TABLE agent_configurations ADD enabled BIT NOT NULL CONSTRAINT DF_agent_configurations_enabled DEFAULT 1;
END

IF EXISTS (SELECT 1 FROM dsc_infra_info WHERE id = 1)
BEGIN
    UPDATE dsc_infra_info SET db_version = '1.2.0', updated_at = GETDATE() WHERE id = 1;
END
ELSE
BEGIN
    INSERT INTO dsc_infra_info (id, web_version, db_version, updated_at)
    VALUES (1, '0.0.1', '1.2.0', GETDATE());
END
