package schema

type Agent struct {
	AgentId                string `json:"agent_id"`
	NodeName               string `json:"node_name"`
	LCMVersion             string `json:"lcm_version"`
	RegistrationType       string `json:"registration_type"`
	CertificateThumbprint  string `json:"certificate_thumbprint"`
	CertificateSubject     string `json:"certificate_subject"`
	CertificateIssuer      string `json:"certificate_issuer"`
	CertificateNotBefore   string `json:"certificate_notbefore"`
	CertificateNotAfter    string `json:"certificate_notafter"`
	RegisteredAt           string `json:"registered_at"`
}
