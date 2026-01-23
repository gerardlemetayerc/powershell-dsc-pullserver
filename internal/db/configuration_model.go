package db

import (
	"database/sql"
	"time"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/utils"
)

// ConfigurationModel représente un modèle de configuration MOF
// property: nom de la propriété
// value: valeur de la propriété
// mofFile: contenu du fichier MOF (BLOB)
// uploadDate: date d'upload
func CreateConfigurationModel(db *sql.DB, cm *schema.ConfigurationModel) error {
	_, err := db.Exec(`INSERT INTO configuration_model (name, uploaded_by, mof_file) VALUES (?, ?, ?)`, cm.Name, cm.UploadedBy, cm.MofFile)
	return err
}

func GetConfigurationModel(db *sql.DB, id int64) (*schema.ConfigurationModel, error) {
	row := db.QueryRow(`SELECT id, name, uploaded_by, mof_file, upload_date, last_usage FROM configuration_model WHERE id = ?`, id)
	var cm schema.ConfigurationModel
	var uploadDate string
	var lastUsage string
	if err := row.Scan(&cm.ID, &cm.Name, &cm.UploadedBy, &cm.MofFile, &uploadDate, &lastUsage); err != nil {
		return nil, err
	}
	// Gestion robuste du parsing de date (supporte format SQLite ISO8601 ou classique)
	cm.UploadDate = utils.ParseTimeFlexible(uploadDate)
	cm.LastUsage = utils.ParseTimeFlexible(lastUsage)
	return &cm, nil
}

func ListConfigurationModels(db *sql.DB) ([]schema.ConfigurationModel, error) {
	rows, err := db.Query(`SELECT id, name, uploaded_by, mof_file, upload_date, last_usage FROM configuration_model ORDER BY upload_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []schema.ConfigurationModel
	for rows.Next() {
		var cm schema.ConfigurationModel
		var uploadDate string
		var lastUsage sql.NullString
		if err := rows.Scan(&cm.ID, &cm.Name, &cm.UploadedBy, &cm.MofFile, &uploadDate, &lastUsage); err != nil {
			return nil, err
		}
		cm.UploadDate = utils.ParseTimeFlexible(uploadDate)
		if lastUsage.Valid && lastUsage.String != "" {
			cm.LastUsage = utils.ParseTimeFlexible(lastUsage.String)
		} else {
			cm.LastUsage = time.Time{}
		}
		list = append(list, cm)
	}
	return list, nil
}

func UpdateConfigurationModel(db *sql.DB, cm *schema.ConfigurationModel) error {
	_, err := db.Exec(`UPDATE configuration_model SET name = ?, uploaded_by = ?, mof_file = ? WHERE id = ?`, cm.Name, cm.UploadedBy, cm.MofFile, cm.ID)
	return err
}

func DeleteConfigurationModel(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM configuration_model WHERE id = ?`, id)
	return err
}
