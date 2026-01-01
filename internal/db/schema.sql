-- Table pour l'historique des rapports DSC
CREATE TABLE IF NOT EXISTS reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    agent_id TEXT,
    job_id TEXT,
    report_format_version TEXT,
    operation_type TEXT,
    refresh_mode TEXT,
    status TEXT,
    start_time TEXT,
    end_time TEXT,
    reboot_requested TEXT,
    errors TEXT,           -- JSON array
    status_data TEXT,      -- JSON array
    additional_data TEXT,  -- JSON array
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    raw_json TEXT
);
CREATE INDEX IF NOT EXISTS idx_reports_agent_id ON reports(agent_id);
-- Table pour suivre les informations d'infrastructure DSC (version web, version db, date MAJ)
CREATE TABLE IF NOT EXISTS dsc_infra_info (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    web_version TEXT DEFAULT '0.0.1',
    db_version TEXT DEFAULT '0.0.2',
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
UPDATE dsc_infra_info SET db_version = '0.0.2', updated_at = CURRENT_TIMESTAMP WHERE id = 1;
-- Schéma pour la table agents

CREATE TABLE IF NOT EXISTS agents (
    agent_id TEXT PRIMARY KEY,
    node_name TEXT,
    lcm_version TEXT,
    registration_type TEXT,
    certificate_thumbprint TEXT,
    certificate_subject TEXT,
    certificate_issuer TEXT,
    certificate_notbefore TEXT,
    certificate_notafter TEXT,
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_communication TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    has_error_last_report BOOLEAN DEFAULT 0
);
-- Migration : ajout de la colonne has_error_last_report si besoin
ALTER TABLE agents ADD COLUMN has_error_last_report BOOLEAN DEFAULT 0;


-- Table de relation 1-n pour les noms de configuration
CREATE TABLE IF NOT EXISTS agent_configurations (
    agent_id TEXT,
    configuration_name TEXT,
    PRIMARY KEY (agent_id, configuration_name),
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);

CREATE TABLE IF NOT EXISTS agent_ips (
    agent_id TEXT,
    ip_address TEXT,
    PRIMARY KEY (agent_id, ip_address),
    FOREIGN KEY (agent_id) REFERENCES agents(agent_id)
);
