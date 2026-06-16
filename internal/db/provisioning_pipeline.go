package db

import (
	"database/sql"
	"errors"
	"fmt"
)

type ProvisioningPipelineConfig struct {
	Enabled       bool
	Provider      string
	APIBaseURL    string
	ProjectPath   string
	WorkflowID    string
	PipelineRef   string
	SecretToken   string
	TimeoutSeconds int
}

type ProvisioningPipelineRun struct {
	AgentID       string
	InternalDSCID *string
	NodeName      *string
	Provider      string
	Status        string
	Message       string
	RemoteRunID   *string
	RemoteURL     *string
	TriggeredBy   string
}

func GetProvisioningPipelineConfig(database *sql.DB) (ProvisioningPipelineConfig, error) {
	var cfg ProvisioningPipelineConfig
	err := database.QueryRow(`
		SELECT enabled, provider, api_base_url, project_path, workflow_id, pipeline_ref, secret_token, timeout_seconds
		FROM provisioning_pipeline_config
		WHERE id = 1`).
		Scan(
			&cfg.Enabled,
			&cfg.Provider,
			&cfg.APIBaseURL,
			&cfg.ProjectPath,
			&cfg.WorkflowID,
			&cfg.PipelineRef,
			&cfg.SecretToken,
			&cfg.TimeoutSeconds,
		)
	if errors.Is(err, sql.ErrNoRows) {
		if _, insErr := database.Exec(`
			INSERT INTO provisioning_pipeline_config (
				id,
				enabled,
				provider,
				api_base_url,
				project_path,
				workflow_id,
				pipeline_ref,
				secret_token,
				timeout_seconds,
				updated_at
			) VALUES (1, 0, 'github', '', '', '', 'main', '', 30, CURRENT_TIMESTAMP)`); insErr != nil {
			return cfg, insErr
		}
		return GetProvisioningPipelineConfig(database)
	}
	return cfg, err
}

func UpsertProvisioningPipelineConfig(database *sql.DB, cfg ProvisioningPipelineConfig) error {
	res, err := database.Exec(`
		UPDATE provisioning_pipeline_config
		SET enabled = ?,
			provider = ?,
			api_base_url = ?,
			project_path = ?,
			workflow_id = ?,
			pipeline_ref = ?,
			secret_token = ?,
			timeout_seconds = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = 1`,
		cfg.Enabled,
		cfg.Provider,
		cfg.APIBaseURL,
		cfg.ProjectPath,
		cfg.WorkflowID,
		cfg.PipelineRef,
		cfg.SecretToken,
		cfg.TimeoutSeconds,
	)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	_, err = database.Exec(`
		INSERT INTO provisioning_pipeline_config (
			id,
			enabled,
			provider,
			api_base_url,
			project_path,
			workflow_id,
			pipeline_ref,
			secret_token,
			timeout_seconds,
			updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)`,
		1,
		cfg.Enabled,
		cfg.Provider,
		cfg.APIBaseURL,
		cfg.ProjectPath,
		cfg.WorkflowID,
		cfg.PipelineRef,
		cfg.SecretToken,
		cfg.TimeoutSeconds,
	)
	return err
}

func InsertProvisioningPipelineRun(database *sql.DB, run ProvisioningPipelineRun) error {
	_, err := database.Exec(`
		INSERT INTO provisioning_pipeline_runs (
			agent_id,
			internal_dsc_id,
			node_name,
			provider,
			status,
			message,
			remote_run_id,
			remote_url,
			triggered_by
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		run.AgentID,
		run.InternalDSCID,
		run.NodeName,
		run.Provider,
		run.Status,
		run.Message,
		run.RemoteRunID,
		run.RemoteURL,
		run.TriggeredBy,
	)
	if err != nil {
		return fmt.Errorf("insert provisioning run: %w", err)
	}
	return nil
}
