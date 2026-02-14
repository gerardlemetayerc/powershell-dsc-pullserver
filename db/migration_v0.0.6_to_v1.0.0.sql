-- Migration script from v0.0.6 to v1.0.0
-- Adds mof_applied column to reports table if not already present

-- Check if column exists, then add if missing
IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'reports' AND COLUMN_NAME = 'mof_applied')
    ALTER TABLE reports ADD mof_applied BIT DEFAULT 0;

UPDATE dsc_infra_info SET db_version = '1.0.0', updated_at = GETDATE() WHERE id = 1;

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
