IF EXISTS (SELECT 1 FROM dsc_infra_info WHERE id = 1)
	UPDATE dsc_infra_info SET db_version = '1.1.4', updated_at = GETDATE() WHERE id = 1;
ELSE
	INSERT INTO dsc_infra_info (id, web_version, db_version, updated_at) VALUES (1, '0.0.1', '1.1.4', GETDATE());
