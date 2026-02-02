package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"io/ioutil"
	"path/filepath"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/utils"
)

// SAMLConfigAPIHandler handles GET and PUT for SAML config
func SAMLConfigAPIHandler(w http.ResponseWriter, r *http.Request) {
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   w.WriteHeader(http.StatusInternalServerError)
		   json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get exe path"})
		   return
	   }
	   configPath := "config.json"
	w.Header().Set("Content-Type", "application/json")
	switch r.Method {
	case http.MethodGet:
		cfg, err := internal.LoadAppConfig("config.json")
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
		_, err := utils.ExePath()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "Failed to get exe path"})
			return
		}
		absPath := filepath.Join(filepath.Dir(exeDir), configPath)
		file, err := os.Open(absPath)
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
