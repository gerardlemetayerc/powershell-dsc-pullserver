package schema

type HTTPConfig struct {
	Port        int    `json:"port"`
	EnableHTTPS bool   `json:"enable_https"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
}

type AppConfig struct {
	Driver        string      `json:"driver"`
	Server        string      `json:"server"`
	Port          int         `json:"port"`
	User          string      `json:"user"`
	Password      string      `json:"password"`
	Database      string      `json:"database"`
	DSCPullServer HTTPConfig  `json:"dsc_pullserver"`
	WebUI         HTTPConfig  `json:"web_ui"`
	SAML          SAMLConfig  `json:"saml"`
}

type SAMLConfig struct {
	Enabled         bool   `json:"enabled"`
	EntityID        string `json:"entity_id"`
	IdpMetadataURL  string `json:"idp_metadata_url"`
	SPKeyFile       string `json:"sp_key_file"`
	SPCertFile      string `json:"sp_cert_file"`
	UserMapping     SAMLAttributeMapping `json:"user_mapping"`
}

type SAMLAttributeMapping struct {
	Email       string `json:"email"`
	Sn          string `json:"sn"`
	GivenName   string `json:"givenname"`
}

type SAMLUserMapping map[string]string