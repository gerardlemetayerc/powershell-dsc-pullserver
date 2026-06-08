package utils

import "strings"

// ExtractConfigName extrait le ConfigurationName d'un segment de type (ConfigurationName='...')
func ExtractConfigName(raw string) string {
	const prefix = "Configurations(ConfigurationName='"
	const suffix = "')"

	if strings.HasPrefix(raw, prefix) && strings.HasSuffix(raw, suffix) {
		s := strings.TrimPrefix(raw, prefix)
		s = strings.TrimSuffix(s, suffix)
		return s
	}

	return raw
}
