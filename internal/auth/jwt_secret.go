package auth

import "go-dsc-pull/internal/schema"

// ResolveJWTSecret returns the JWT signing secret used by web authentication.
// Priority is web_ui.shared_secret, then dsc_pullserver.shared_secret for backward compatibility.
func ResolveJWTSecret(appCfg *schema.AppConfig) string {
	if appCfg == nil {
		return ""
	}
	if appCfg.WebUI.SharedAccessSecret != "" {
		return appCfg.WebUI.SharedAccessSecret
	}
	return appCfg.DSCPullServer.SharedAccessSecret
}