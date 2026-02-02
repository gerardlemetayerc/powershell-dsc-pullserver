package schema

type HTTPConfig struct {
	Port        int    `json:"port"`
	EnableHTTPS bool   `json:"enable_https"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
	EnableClientCertValidation bool `json:"enable_client_cert_validation"`
	BypassCAValidation bool `json:"bypass_ca_validation"`
	RegistrationKey string `json:"registrationKey"`
	SharedAccessSecret string `json:"shared_secret"`
}

type AppConfig struct {
	Driver        string      `json:"driver"`
	Server        string      `json:"server"`
	Port          int         `json:"port"`
	User          string      `json:"user"`
	Password      string      `json:"password"`
	Database      DBConfig    `json:"database"`
	DSCPullServer HTTPConfig  `json:"dsc_pullserver"`
	WebUI         HTTPConfig  `json:"web_ui"`
	SAML          SAMLConfig  `json:"saml"`
}

type DBConfig struct {
	Driver   string `json:"driver"`
	Server   string `json:"server"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"name"`
}

type SAMLConfig struct {
	Enabled         bool   `json:"enabled"`
	EntityID        string `json:"entity_id"`
	IdpMetadataURL  string `json:"idp_metadata_url"`
	SPKeyFile       string `json:"sp_key_file"`
	SPCertFile      string `json:"sp_cert_file"`
	UserMapping     SAMLAttributeMapping `json:"user_mapping"`
	GroupMapping    SAMLGroupMapping     `json:"group_mapping"`
}

type SAMLAttributeMapping struct {
	Email       string `json:"email"`
	Sn          string `json:"sn"`
	GivenName   string `json:"givenname"`
}

type SAMLGroupMapping struct {
	Attribute   string `json:"attribute"`
	AdminValue  string `json:"admin_value"`
	UserValue   string `json:"user_value"`
}

type SAMLUserMapping map[string]string