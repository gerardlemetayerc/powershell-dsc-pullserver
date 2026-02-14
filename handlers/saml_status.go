package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal/global"
)

// SAMLStatusHandler retourne si SAML est activé
func SAMLStatusHandler(w http.ResponseWriter, r *http.Request) {
	   cfg := global.AppConfig
	   if cfg == nil {
		   w.WriteHeader(http.StatusInternalServerError)
		   json.NewEncoder(w).Encode(map[string]bool{"enabled": false})
		   return
	   }
	   json.NewEncoder(w).Encode(map[string]bool{"enabled": cfg.SAML.Enabled})
}
