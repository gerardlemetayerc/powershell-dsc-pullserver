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
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

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
