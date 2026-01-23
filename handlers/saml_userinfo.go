package handlers

import (
	"net/http"
	"github.com/crewjam/saml/samlsp"
	"go-dsc-pull/internal"
	"encoding/json"
)

// SAMLUserInfoHandler : expose les infos utilisateur extraites des claims SAML selon le mapping
func SAMLUserInfoHandler(w http.ResponseWriter, r *http.Request) {
	sess := samlsp.SessionFromContext(r.Context())
	if sess == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "no SAML session"})
		return
	}
	mapping, _ := internal.GetSAMLUserMapping()
	attrs := map[string]string{}
	if custom, ok := sess.(interface{ GetAttributes() map[string][]string }); ok {
		for samlClaim, localField := range mapping {
			values := custom.GetAttributes()[samlClaim]
			if len(values) > 0 {
				attrs[localField] = values[0]
			}
		}
	}
	json.NewEncoder(w).Encode(attrs)
}
