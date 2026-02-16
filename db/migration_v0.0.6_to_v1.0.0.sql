-- Migration script from v0.0.6 to v1.0.0
-- Adds mof_applied column to reports table if not already present

-- Check if column exists, then add if missing
IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'reports' AND COLUMN_NAME = 'mof_applied')
    ALTER TABLE reports ADD mof_applied BIT DEFAULT 0;

UPDATE dsc_infra_info SET db_version = '1.0.0', updated_at = GETDATE() WHERE id = 1;

-- Add recommended indexes for performance
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_configuration_model_name' AND object_id = OBJECT_ID('configuration_model'))
    CREATE INDEX idx_configuration_model_name ON configuration_model(name);

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_configuration_model_original_name' AND object_id = OBJECT_ID('configuration_model'))
    CREATE INDEX idx_configuration_model_original_name ON configuration_model(original_name);

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_agent_configurations_configuration_name' AND object_id = OBJECT_ID('agent_configurations'))
    CREATE INDEX idx_agent_configurations_configuration_name ON agent_configurations(configuration_name);

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_agents_state' AND object_id = OBJECT_ID('agents'))
    CREATE INDEX idx_agents_state ON agents(state);

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_reports_job_id' AND object_id = OBJECT_ID('reports'))
    CREATE INDEX idx_reports_job_id ON reports(job_id);

IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_modules_name' AND object_id = OBJECT_ID('modules'))
    CREATE INDEX idx_modules_name ON modules(name);

-- Ajout table d'audit pour tracer les actions utilisateurs
IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='audit' AND xtype='U')
CREATE TABLE audit (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_email NVARCHAR(255) NULL,
    action NVARCHAR(128) NOT NULL,
    target NVARCHAR(255) NULL,
    details NVARCHAR(MAX) NULL,
    ip_address NVARCHAR(45) NULL,
    created_at DATETIME DEFAULT GETDATE()
);
