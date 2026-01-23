package utils

import (
	"strings"
)

// ExtractAgentId extrait l'AgentId d'un segment de type (AgentId='...')
func ExtractAgentId(raw string) string {
	// Supporte (AgentId='...') ou Nodes(AgentId='...')
	start := strings.Index(raw, "(AgentId='")
	end := strings.LastIndex(raw, "')")
	if start != -1 && end != -1 && end > start+10 {
		return raw[start+10 : end]
	}
	return ""
}
