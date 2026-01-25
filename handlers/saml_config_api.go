package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"io/ioutil"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/schema"
)

// SAMLConfigAPIHandler handles GET and PUT for SAML config
func SAMLConfigAPIHandler(w http.ResponseWriter, r *http.Request) {
	configPath := "config.json"
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		cfg, err := internal.LoadAppConfig(configPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to load config"})
			return
		}
		json.NewEncoder(w).Encode(cfg.SAML)
	case http.MethodPut:
		// Only allow admin (add your own middleware/check here)
		var newSAML schema.SAMLConfig
		if err := json.NewDecoder(r.Body).Decode(&newSAML); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Invalid JSON"})
			return
		}
		// Load full config
		file, err := os.Open(configPath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to open config"})
			return
		}
		defer file.Close()
		var fullConfig map[string]interface{}
		if err := json.NewDecoder(file).Decode(&fullConfig); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to decode config"})
			return
		}
		// Overwrite SAML config
		fullConfig["saml"] = newSAML
		// Write back to file
		data, err := json.MarshalIndent(fullConfig, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to marshal config"})
			return
		}
		if err := ioutil.WriteFile(configPath, data, 0644); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to write config"})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
