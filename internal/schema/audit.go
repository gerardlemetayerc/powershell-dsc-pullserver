package schema

// Audit représente une entrée d'audit utilisateur
type Audit struct {
	ID         int64  // id en base
	UserEmail  string // email de l'utilisateur (nullable)
	Action     string // type d'action
	Target     string // cible de l'action (nullable)
	Details    string // détails ou JSON (nullable)
	IPAddress  string // adresse IP (nullable)
	CreatedAt  string // timestamp
}
