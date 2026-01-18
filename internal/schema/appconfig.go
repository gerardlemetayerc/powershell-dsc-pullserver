package schema

type AppConfig struct {
	Driver   string      `json:"driver"`
	Server   string      `json:"server"`
	Port     int         `json:"port"`
	User     string      `json:"user"`
	Password string      `json:"password"`
	Database string      `json:"database"`
	DSCPort  int         `json:"dsc_port"`
	WebPort  int         `json:"web_port"`
	SAML     SAMLConfig  `json:"saml"`
}

type SAMLConfig struct {
	Enabled         bool   `json:"enabled"`
	EntityID        string `json:"entity_id"`
	ACSURL          string `json:"acs_url"`
	IdpMetadataURL  string `json:"idp_metadata_url"`
	SPKeyFile       string `json:"sp_key_file"`
	SPCertFile      string `json:"sp_cert_file"`
}

type SAMLUserMapping map[string]string