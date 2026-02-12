package schema

import "time"

// ConfigurationModel représente un modèle de configuration MOF
// property: nom de la propriété
// value: valeur de la propriété
// mofFile: contenu du fichier MOF (BLOB)
// uploadDate: date d'upload
type ConfigurationModel struct {
	ID           int64       `json:"id"`
	Name         string      `json:"name"`
	OriginalName *string     `json:"original_name,omitempty"`
	PreviousID   *int64      `json:"previous_id,omitempty"`
	UploadDate   time.Time   `json:"upload_date"`
	UploadedBy   string      `json:"uploaded_by"`
	MofFile      []byte      `json:"mof_file"`
	LastUsage    *time.Time  `json:"last_usage,omitempty"`
}
