package db

import (
	"database/sql"
	"time"
)

// ConfigurationModel représente un modèle de configuration MOF
// property: nom de la propriété
// value: valeur de la propriété
// mofFile: contenu du fichier MOF (BLOB)
// uploadDate: date d'upload

type ConfigurationModel struct {
	ID         int64     `json:"id"`
	Property   string    `json:"property"`
	Value      string    `json:"value"`
	MofFile    []byte    `json:"mof_file"`
	UploadDate time.Time `json:"upload_date"`
}

func CreateConfigurationModel(db *sql.DB, cm *ConfigurationModel) error {
	_, err := db.Exec(`INSERT INTO configuration_model (property, value, mof_file) VALUES (?, ?, ?)`, cm.Property, cm.Value, cm.MofFile)
	return err
}

func GetConfigurationModel(db *sql.DB, id int64) (*ConfigurationModel, error) {
	row := db.QueryRow(`SELECT id, property, value, mof_file, upload_date FROM configuration_model WHERE id = ?`, id)
	var cm ConfigurationModel
	var uploadDate string
	if err := row.Scan(&cm.ID, &cm.Property, &cm.Value, &cm.MofFile, &uploadDate); err != nil {
		return nil, err
	}
	d, _ := time.Parse("2006-01-02 15:04:05", uploadDate)
	cm.UploadDate = d
	return &cm, nil
}

func ListConfigurationModels(db *sql.DB) ([]ConfigurationModel, error) {
	rows, err := db.Query(`SELECT id, property, value, mof_file, upload_date FROM configuration_model ORDER BY upload_date DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ConfigurationModel
	for rows.Next() {
		var cm ConfigurationModel
		var uploadDate string
		if err := rows.Scan(&cm.ID, &cm.Property, &cm.Value, &cm.MofFile, &uploadDate); err != nil {
			return nil, err
		}
		d, _ := time.Parse("2006-01-02 15:04:05", uploadDate)
		cm.UploadDate = d
		list = append(list, cm)
	}
	return list, nil
}

func UpdateConfigurationModel(db *sql.DB, cm *ConfigurationModel) error {
	_, err := db.Exec(`UPDATE configuration_model SET property = ?, value = ?, mof_file = ? WHERE id = ?`, cm.Property, cm.Value, cm.MofFile, cm.ID)
	return err
}

func DeleteConfigurationModel(db *sql.DB, id int64) error {
	_, err := db.Exec(`DELETE FROM configuration_model WHERE id = ?`, id)
	return err
}
