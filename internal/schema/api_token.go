package schema

type APIToken struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	Label     string `json:"label"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	RevokedAt *string `json:"revoked_at"`
}
