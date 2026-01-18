package internal

// SAMLConfig structure pour la configuration SAML
// À compléter avec les valeurs de votre IdP
// Exemple à intégrer dans config.json
// "saml": {
//   "enabled": true,
//   "entity_id": "https://votre-serveur/saml/metadata",
//   "acs_url": "https://votre-serveur/saml/acs",
//   "idp_metadata_url": "https://idp.example.com/metadata",
//   "sp_key_file": "sp.key",
//   "sp_cert_file": "sp.crt"
// }
type SAMLConfig struct {
	Enabled         bool   `json:"enabled"`
	EntityID        string `json:"entity_id"`
	ACSURL          string `json:"acs_url"`
	IdpMetadataURL  string `json:"idp_metadata_url"`
	SPKeyFile       string `json:"sp_key_file"`
	SPCertFile      string `json:"sp_cert_file"`
}
