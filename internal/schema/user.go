package schema

type User struct {
	ID             int64   `json:"id"`
	FirstName      string  `json:"first_name"`
	LastName       string  `json:"last_name"`
	Email          string  `json:"email"`
	PasswordHash   string  `json:"password_hash,omitempty"`
	Role           string  `json:"role"`
	IsActive       bool    `json:"is_active"`
	CreatedAt      string  `json:"created_at"`
	LastLogonDate  *string `json:"last_logon_date"`
	Source         string  `json:"source"`
}
