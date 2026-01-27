package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal"
	"path/filepath"
	"go-dsc-pull/utils"
)

// SAMLStatusHandler retourne si SAML est activé
func SAMLStatusHandler(w http.ResponseWriter, r *http.Request) {
	   exeDir, err := utils.ExePath()
	   if err != nil {
		   w.WriteHeader(http.StatusInternalServerError)
		   json.NewEncoder(w).Encode(map[string]bool{"enabled": false})
		   return
	   }
	   configPath := filepath.Join(filepath.Dir(exeDir), "config.json")
	   cfg, err := internal.LoadAppConfig(configPath)
	   if err != nil {
		   w.WriteHeader(http.StatusInternalServerError)
		   json.NewEncoder(w).Encode(map[string]bool{"enabled": false})
		   return
	   }
	   json.NewEncoder(w).Encode(map[string]bool{"enabled": cfg.SAML.Enabled})
}
