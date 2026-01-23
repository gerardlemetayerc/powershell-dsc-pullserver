package utils

import "strings"

// ExtractConfigName extrait le ConfigurationName d'un segment de type (ConfigurationName='...')
func ExtractConfigName(raw string) string {
	s := strings.TrimPrefix(raw, "Configurations(ConfigurationName='") 
	s = strings.TrimSuffix(s, "')")
	return s
}
