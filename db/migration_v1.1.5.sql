IF EXISTS (SELECT * FROM sysobjects WHERE name='agents' AND xtype='U')
BEGIN
	IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agents' AND COLUMN_NAME = 'internal_dsc_id')
		ALTER TABLE agents ADD internal_dsc_id NVARCHAR(128) NULL;

	IF EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'agents' AND COLUMN_NAME = 'internal_dsc_id')
	BEGIN
		IF EXISTS (SELECT 1 FROM sys.indexes WHERE name = 'idx_agents_internal_dsc_id' AND object_id = OBJECT_ID('agents') AND has_filter = 1)
			EXEC sp_executesql N'DROP INDEX idx_agents_internal_dsc_id ON agents';

		EXEC sp_executesql N'UPDATE agents SET internal_dsc_id = CONCAT(''IDSC-'', REPLACE(CONVERT(NVARCHAR(36), NEWID()), ''-'', '''')) WHERE internal_dsc_id IS NULL';

		IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_agents_internal_dsc_id' AND object_id = OBJECT_ID('agents'))
			EXEC sp_executesql N'CREATE UNIQUE INDEX idx_agents_internal_dsc_id ON agents(internal_dsc_id)';
	END
END