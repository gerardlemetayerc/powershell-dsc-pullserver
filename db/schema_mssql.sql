IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='users' AND xtype='U')
CREATE TABLE users (
    id INT IDENTITY(1,1) PRIMARY KEY,
    first_name NVARCHAR(128) NOT NULL,
    last_name NVARCHAR(128) NOT NULL,
    email NVARCHAR(255) NOT NULL UNIQUE,
    password_hash NVARCHAR(255) NOT NULL,
    is_active BIT DEFAULT 1,
    created_at DATETIME DEFAULT GETDATE(),
    last_logon_date DATETIME,
    role NVARCHAR(50) DEFAULT 'readonly',
    source NVARCHAR(50) DEFAULT 'local'
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='user_api_tokens' AND xtype='U')
CREATE TABLE user_api_tokens (
    id INT IDENTITY(1,1) PRIMARY KEY,
    user_id INT NOT NULL,
    token_hash NVARCHAR(255) NOT NULL,
    label NVARCHAR(255),
    is_active BIT DEFAULT 1,
    created_at DATETIME DEFAULT GETDATE(),
    revoked_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='modules' AND xtype='U')
CREATE TABLE modules (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(255) NOT NULL,
    version NVARCHAR(50) NOT NULL,
    checksum NVARCHAR(255) NOT NULL,
    zip_blob VARBINARY(MAX) NOT NULL,
    uploaded_at DATETIME DEFAULT GETDATE()
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='reports' AND xtype='U')
CREATE TABLE reports (
    id INT IDENTITY(1,1) PRIMARY KEY,
    agent_id NVARCHAR(128),
    job_id NVARCHAR(128),
    report_format_version NVARCHAR(50),
    operation_type NVARCHAR(50),
    refresh_mode NVARCHAR(50),
    status NVARCHAR(50),
    start_time NVARCHAR(50),
    end_time NVARCHAR(50),
    reboot_requested NVARCHAR(10),
    errors NVARCHAR(MAX),           -- JSON array
    status_data NVARCHAR(MAX),      -- JSON array
    additional_data NVARCHAR(MAX),  -- JSON array
    created_at DATETIME DEFAULT GETDATE(),
    raw_json NVARCHAR(MAX)
);
IF NOT EXISTS (SELECT * FROM sys.indexes WHERE name = 'idx_reports_agent_id' AND object_id = OBJECT_ID('reports'))
    CREATE INDEX idx_reports_agent_id ON reports(agent_id);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='dsc_infra_info' AND xtype='U')
CREATE TABLE dsc_infra_info (
    id INT PRIMARY KEY CHECK (id = 1),
    web_version NVARCHAR(20) DEFAULT '0.0.1',
    db_version NVARCHAR(20) DEFAULT '0.0.4',
    updated_at DATETIME DEFAULT GETDATE()
);
-- To update version:
-- UPDATE dsc_infra_info SET db_version = '0.0.4', updated_at = GETDATE() WHERE id = 1;

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='agents' AND xtype='U')
CREATE TABLE agents (
    agent_id NVARCHAR(128) PRIMARY KEY,
    node_name NVARCHAR(128),
    lcm_version NVARCHAR(50) NULL,
    registration_type NVARCHAR(50) NULL,
    certificate_thumbprint NVARCHAR(128) NULL,
    certificate_subject NVARCHAR(255) NULL,
    certificate_issuer NVARCHAR(255) NULL,
    certificate_notbefore NVARCHAR(50) NULL,
    certificate_notafter NVARCHAR(50) NULL,
    registered_at DATETIME DEFAULT GETDATE(),
    last_communication DATETIME DEFAULT GETDATE(),
    state NVARCHAR(50),
    has_error_last_report BIT DEFAULT 0
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='agent_configurations' AND xtype='U')
CREATE TABLE agent_configurations (
    agent_id NVARCHAR(128),
    configuration_name NVARCHAR(128),
    PRIMARY KEY (agent_id, configuration_name),
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='agent_ips' AND xtype='U')
CREATE TABLE agent_ips (
    agent_id NVARCHAR(128),
    ip_address NVARCHAR(45),
    PRIMARY KEY (agent_id, ip_address),
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='properties' AND xtype='U')
CREATE TABLE properties (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(128) NOT NULL UNIQUE,
    description NVARCHAR(255),
    priority INT DEFAULT 0
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='node_properties' AND xtype='U')
CREATE TABLE node_properties (
    node_id NVARCHAR(128) NOT NULL,
    property_id INT NOT NULL,
    value NVARCHAR(255),
    PRIMARY KEY (node_id, property_id),
    FOREIGN KEY (property_id) REFERENCES properties(id)
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='configuration_model' AND xtype='U')
CREATE TABLE configuration_model (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(128) NOT NULL,
    upload_date DATETIME DEFAULT GETDATE(),
    uploaded_by NVARCHAR(128) NOT NULL,
    mof_file VARBINARY(MAX) NOT NULL,
    last_usage DATETIME
);

IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='agent_tags' AND xtype='U')
CREATE TABLE agent_tags (
    agent_id NVARCHAR(128) NOT NULL,
    tag_key NVARCHAR(128) NOT NULL,
    tag_value NVARCHAR(255) NOT NULL,
    PRIMARY KEY (agent_id, tag_key, tag_value),
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);