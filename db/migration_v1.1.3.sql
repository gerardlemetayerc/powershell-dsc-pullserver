IF EXISTS (SELECT 1 FROM dsc_infra_info WHERE id = 1)
	UPDATE dsc_infra_info SET db_version = '1.1.3', updated_at = GETDATE() WHERE id = 1;
ELSE
	INSERT INTO dsc_infra_info (id, web_version, db_version, updated_at) VALUES (1, '0.0.1', '1.1.3', GETDATE());

IF EXISTS (SELECT * FROM sysobjects WHERE name='dsc_infra_info' AND xtype='U')
BEGIN
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'dsc_infra_info' AND COLUMN_NAME = 'latest_release')
		ALTER TABLE dsc_infra_info ADD latest_release NVARCHAR(50) NULL;
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'dsc_infra_info' AND COLUMN_NAME = 'latest_release_url')
		ALTER TABLE dsc_infra_info ADD latest_release_url NVARCHAR(255) NULL;
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'dsc_infra_info' AND COLUMN_NAME = 'update_available')
		ALTER TABLE dsc_infra_info ADD update_available BIT DEFAULT 0;
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'dsc_infra_info' AND COLUMN_NAME = 'release_check_ok')
		ALTER TABLE dsc_infra_info ADD release_check_ok BIT DEFAULT 0;
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'dsc_infra_info' AND COLUMN_NAME = 'release_checked_at')
		ALTER TABLE dsc_infra_info ADD release_checked_at DATETIME NULL;
END
