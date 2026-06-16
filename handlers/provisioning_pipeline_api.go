package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/global"
)

type provisioningPipelineConfigResponse struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	APIBaseURL     string `json:"api_base_url"`
	ProjectPath    string `json:"project_path"`
	WorkflowID     string `json:"workflow_id"`
	PipelineRef    string `json:"pipeline_ref"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	HasSecret      bool   `json:"has_secret"`
}

type provisioningPipelineConfigUpdateRequest struct {
	Enabled        bool    `json:"enabled"`
	Provider       string  `json:"provider"`
	APIBaseURL     string  `json:"api_base_url"`
	ProjectPath    string  `json:"project_path"`
	WorkflowID     string  `json:"workflow_id"`
	PipelineRef    string  `json:"pipeline_ref"`
	TimeoutSeconds int     `json:"timeout_seconds"`
	SecretToken    *string `json:"secret_token,omitempty"`
}

type provisioningRunDispatchResult struct {
	RemoteRunID *string
	RemoteURL   *string
	Message     string
}

type provisioningRunRequest struct {
	Inputs map[string]string `json:"inputs"`
}

func ProvisioningPipelineConfigAPIHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		driverName := global.AppConfig.Database.Driver

		switch r.Method {
		case http.MethodGet:
			cfg, err := db.GetProvisioningPipelineConfig(database)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load provisioning config"})
				return
			}
			_ = json.NewEncoder(w).Encode(provisioningPipelineConfigResponse{
				Enabled:        cfg.Enabled,
				Provider:       cfg.Provider,
				APIBaseURL:     cfg.APIBaseURL,
				ProjectPath:    cfg.ProjectPath,
				WorkflowID:     cfg.WorkflowID,
				PipelineRef:    cfg.PipelineRef,
				TimeoutSeconds: cfg.TimeoutSeconds,
				HasSecret:      strings.TrimSpace(cfg.SecretToken) != "",
			})
		case http.MethodPut:
			var req provisioningPipelineConfigUpdateRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
				return
			}

			provider := strings.ToLower(strings.TrimSpace(req.Provider))
			if provider != "github" && provider != "gitlab" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "provider must be github or gitlab"})
				return
			}

			cfg, err := db.GetProvisioningPipelineConfig(database)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load existing config"})
				return
			}

			cfg.Enabled = req.Enabled
			cfg.Provider = provider
			cfg.APIBaseURL = strings.TrimSpace(req.APIBaseURL)
			cfg.ProjectPath = strings.TrimSpace(req.ProjectPath)
			cfg.WorkflowID = strings.TrimSpace(req.WorkflowID)
			cfg.PipelineRef = strings.TrimSpace(req.PipelineRef)
			cfg.TimeoutSeconds = req.TimeoutSeconds
			if cfg.TimeoutSeconds <= 0 {
				cfg.TimeoutSeconds = 30
			}
			if cfg.PipelineRef == "" {
				cfg.PipelineRef = "main"
			}
			if req.SecretToken != nil {
				newSecret := strings.TrimSpace(*req.SecretToken)
				if newSecret != "" {
					cfg.SecretToken = newSecret
				}
			}

			if cfg.Enabled {
				if cfg.APIBaseURL == "" || cfg.ProjectPath == "" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "api_base_url and project_path are required when enabled"})
					return
				}
				if cfg.Provider == "github" && cfg.WorkflowID == "" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "workflow_id is required for github"})
					return
				}
				if strings.TrimSpace(cfg.SecretToken) == "" {
					w.WriteHeader(http.StatusBadRequest)
					_ = json.NewEncoder(w).Encode(map[string]string{"error": "secret_token is required when enabled"})
					return
				}
			}

			if err := db.UpsertProvisioningPipelineConfig(database, cfg); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to persist provisioning config"})
				return
			}

			actor := resolveAuditActor(database, r)
			detail := fmt.Sprintf("Updated provisioning pipeline config (enabled=%t provider=%s)", cfg.Enabled, cfg.Provider)
			_ = db.InsertAudit(database, driverName, actor, "update", "provisioning_pipeline_config", detail, "")

			_ = json.NewEncoder(w).Encode(provisioningPipelineConfigResponse{
				Enabled:        cfg.Enabled,
				Provider:       cfg.Provider,
				APIBaseURL:     cfg.APIBaseURL,
				ProjectPath:    cfg.ProjectPath,
				WorkflowID:     cfg.WorkflowID,
				PipelineRef:    cfg.PipelineRef,
				TimeoutSeconds: cfg.TimeoutSeconds,
				HasSecret:      strings.TrimSpace(cfg.SecretToken) != "",
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func ProvisioningPipelineRunHandler(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		agentID := strings.TrimSpace(r.PathValue("id"))
		if agentID == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Agent id is required"})
			return
		}

		var req provisioningRunRequest
		if r.Body != nil {
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&req); err != nil && err != io.EOF {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON body"})
				return
			}
		}
		customInputs := map[string]string{}
		for k, v := range req.Inputs {
			key := strings.TrimSpace(k)
			if key == "" {
				continue
			}
			customInputs[key] = strings.TrimSpace(v)
		}

		cfg, err := db.GetProvisioningPipelineConfig(database)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load provisioning config"})
			return
		}
		if !cfg.Enabled {
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Provisioning pipeline is disabled"})
			return
		}

		var dbAgentID string
		var internalDSCID *string
		var nodeName *string
		err = database.QueryRow(`SELECT agent_id, internal_dsc_id, node_name FROM agents WHERE agent_id = ?`, agentID).Scan(&dbAgentID, &internalDSCID, &nodeName)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "Agent not found"})
			return
		}

		result, dispatchErr := dispatchProvisioningRun(cfg, dbAgentID, internalDSCID, nodeName, customInputs)
		actor := resolveAuditActor(database, r)

		runStatus := "dispatched"
		runMessage := "Pipeline dispatched"
		if dispatchErr != nil {
			runStatus = "error"
			runMessage = dispatchErr.Error()
		}

		run := db.ProvisioningPipelineRun{
			AgentID:       dbAgentID,
			InternalDSCID: internalDSCID,
			NodeName:      nodeName,
			Provider:      cfg.Provider,
			Status:        runStatus,
			Message:       runMessage,
			RemoteRunID:   result.RemoteRunID,
			RemoteURL:     result.RemoteURL,
			TriggeredBy:   actor,
		}
		_ = db.InsertProvisioningPipelineRun(database, run)

		driverName := global.AppConfig.Database.Driver
		auditDetails := fmt.Sprintf("Provisioning pipeline run for agent_id=%s status=%s", dbAgentID, runStatus)
		_ = db.InsertAudit(database, driverName, actor, "run", "provisioning_pipeline", auditDetails, "")

		if dispatchErr != nil {
			w.WriteHeader(http.StatusBadGateway)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": dispatchErr.Error()})
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "dispatched",
			"provider":   cfg.Provider,
			"message":    result.Message,
			"remote_run": result.RemoteRunID,
			"remote_url": result.RemoteURL,
			"inputs":     customInputs,
		})
	}
}

func dispatchProvisioningRun(cfg db.ProvisioningPipelineConfig, agentID string, internalDSCID *string, nodeName *string, customInputs map[string]string) (provisioningRunDispatchResult, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "github":
		return dispatchGitHubWorkflow(cfg, agentID, internalDSCID, nodeName, customInputs)
	case "gitlab":
		return dispatchGitLabPipeline(cfg, agentID, internalDSCID, nodeName, customInputs)
	default:
		return provisioningRunDispatchResult{}, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}

func dispatchGitHubWorkflow(cfg db.ProvisioningPipelineConfig, agentID string, internalDSCID *string, nodeName *string, customInputs map[string]string) (provisioningRunDispatchResult, error) {
	if strings.TrimSpace(cfg.APIBaseURL) == "" || strings.TrimSpace(cfg.ProjectPath) == "" || strings.TrimSpace(cfg.WorkflowID) == "" || strings.TrimSpace(cfg.SecretToken) == "" {
		return provisioningRunDispatchResult{}, fmt.Errorf("incomplete github configuration")
	}

	inputs := map[string]string{}
	for k, v := range customInputs {
		inputs[k] = v
	}
	// Core identifiers are always enforced from server-side data.
	inputs["agent_id"] = agentID
	if internalDSCID != nil {
		inputs["internal_dsc_id"] = *internalDSCID
	}
	if nodeName != nil {
		inputs["node_name"] = *nodeName
	}

	payload := map[string]interface{}{
		"ref":    cfg.PipelineRef,
		"inputs": inputs,
	}
	body, _ := json.Marshal(payload)

	baseURL := strings.TrimRight(cfg.APIBaseURL, "/")
	workflowID := url.PathEscape(cfg.WorkflowID)
	dispatchURL := fmt.Sprintf("%s/repos/%s/actions/workflows/%s/dispatches", baseURL, strings.Trim(cfg.ProjectPath, "/"), workflowID)

	req, err := http.NewRequest(http.MethodPost, dispatchURL, bytes.NewReader(body))
	if err != nil {
		return provisioningRunDispatchResult{}, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.SecretToken)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("Content-Type", "application/json")

	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}
	resp, err := (&http.Client{Timeout: time.Duration(timeout) * time.Second}).Do(req)
	if err != nil {
		return provisioningRunDispatchResult{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		data, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return provisioningRunDispatchResult{}, fmt.Errorf("github dispatch failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(data)))
	}

	var remoteURL *string
	if strings.Contains(baseURL, "api.github.com") {
		u := fmt.Sprintf("https://github.com/%s/actions", strings.Trim(cfg.ProjectPath, "/"))
		remoteURL = &u
	}

	msg := "GitHub workflow dispatched"
	return provisioningRunDispatchResult{Message: msg, RemoteURL: remoteURL}, nil
}

func dispatchGitLabPipeline(cfg db.ProvisioningPipelineConfig, agentID string, internalDSCID *string, nodeName *string, customInputs map[string]string) (provisioningRunDispatchResult, error) {
	if strings.TrimSpace(cfg.APIBaseURL) == "" || strings.TrimSpace(cfg.ProjectPath) == "" || strings.TrimSpace(cfg.SecretToken) == "" {
		return provisioningRunDispatchResult{}, fmt.Errorf("incomplete gitlab configuration")
	}

	baseURL := strings.TrimRight(cfg.APIBaseURL, "/")	
	projectPath := url.QueryEscape(strings.Trim(cfg.ProjectPath, "/"))
	dispatchURL := fmt.Sprintf("%s/projects/%s/trigger/pipeline", baseURL, projectPath)

	form := url.Values{}
	form.Set("token", cfg.SecretToken)
	form.Set("ref", cfg.PipelineRef)
	for k, v := range customInputs {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		form.Set(fmt.Sprintf("variables[%s]", key), v)
	}
	// Core identifiers are always enforced from server-side data.
	form.Set("variables[AGENT_ID]", agentID)
	if internalDSCID != nil {
		form.Set("variables[INTERNAL_DSC_ID]", *internalDSCID)
	}
	if nodeName != nil {
		form.Set("variables[NODE_NAME]", *nodeName)
	}

	req, err := http.NewRequest(http.MethodPost, dispatchURL, strings.NewReader(form.Encode()))
	if err != nil {
		return provisioningRunDispatchResult{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	timeout := cfg.TimeoutSeconds
	if timeout <= 0 {
		timeout = 30
	}
	resp, err := (&http.Client{Timeout: time.Duration(timeout) * time.Second}).Do(req)
	if err != nil {
		return provisioningRunDispatchResult{}, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return provisioningRunDispatchResult{}, fmt.Errorf("gitlab dispatch failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var parsed map[string]interface{}
	_ = json.Unmarshal(respBody, &parsed)

	var remoteRunID *string
	if idVal, ok := parsed["id"]; ok {
		s := fmt.Sprintf("%v", idVal)
		remoteRunID = &s
	}
	var remoteURL *string
	if uVal, ok := parsed["web_url"].(string); ok && strings.TrimSpace(uVal) != "" {
		remoteURL = &uVal
	}

	msg := "GitLab pipeline dispatched"
	return provisioningRunDispatchResult{Message: msg, RemoteRunID: remoteRunID, RemoteURL: remoteURL}, nil
}

