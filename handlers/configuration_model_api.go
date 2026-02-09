package handlers

import (
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"fmt"
	"log"
	"go-dsc-pull/internal/auth"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal"
	"go-dsc-pull/internal/schema"
	"go-dsc-pull/internal/logs"
	jwt "github.com/golang-jwt/jwt/v5"
)

// POST /api/v1/configuration_models
func CreateConfigurationModelHandler(w http.ResponseWriter, r *http.Request) {
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB config error"))
		return
	}
	dbConn, err := db.OpenDB(dbCfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("DB open error"))
		return
	}
	defer dbConn.Close()
	file, header, err := r.FormFile("mof_file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing MOF file"))
		return
	}
	defer file.Close()
	// Vérifie l'extension .mof
	filename := strings.ToLower(header.Filename)
	log.Printf("[REGISTER][CONFIG] Uploading configuration model: %s", filename)
	if len(filename) < 5 || filename[len(filename)-4:] != ".mof" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Le fichier doit être au format .mof"))
		return
	}
		// Calcule le nom du modèle (nom du fichier sans .mof)
		name := strings.ToLower(filename[:len(filename)-4])
		log.Printf("[REGISTER][CONFIG] Base name extracted: %s", name)
		mofBytes, err := ioutil.ReadAll(file)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// 1. Récupérer la configuration ayant le même nom (si il y en a une), changer son nom "champ name" avec le nom "versionné", et stocker l'ID de cette configuration en variable
		var previousID *int64 = nil
		var existingId int64
		var existingDate string
		var originalName sql.NullString
		err = dbConn.QueryRow(`SELECT id, upload_date, original_name FROM configuration_model WHERE name = ?`, name).Scan(&existingId, &existingDate, &originalName)
		var newName string
		if err == nil && existingId > 0 {
			safeDate := existingDate
			safeDate = strings.ReplaceAll(safeDate, ":", "")
			safeDate = strings.ReplaceAll(safeDate, "-", "")
			safeDate = strings.ReplaceAll(safeDate, " ", "_")
			newName = name + "-" + safeDate
			_, errUpdate := dbConn.Exec("UPDATE configuration_model SET name = ? WHERE id = ?", newName, existingId)
			if errUpdate != nil {
				logs.WriteLogFile("[ERROR] Erreur UPDATE configuration_model versioning: " + errUpdate.Error())
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Versioning error"))
				return
			}
			previousID = &existingId
		}

		// 2. Tenter la création de la configuration, et indiquer si besoin l'ID de la précédente configuration (celle modifiée précédemment) si besoin dans le champ prévu à cet effet
		uploadedBy := "?"
		if r.Context().Value("user") != nil {
			if sub, ok := r.Context().Value("user").(string); ok {
				uploadedBy = sub
			}
		} else if auth := r.Header.Get("Authorization"); len(auth) > 7 {
			tokenStr := auth[7:]
			appCfg, err := internal.LoadAppConfig("config.json")
			if err != nil {
				log.Printf("[REGISTER][CONFIG] Error loading config: %v", err)
				http.Error(w, "Server configuration error: unable to load config", http.StatusInternalServerError)
				return
			}
			secret := []byte(appCfg.DSCPullServer.SharedAccessSecret)
			token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return secret, nil })
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				if sub, ok := claims["sub"].(string); ok {
					uploadedBy = sub
				}
			}
		}

		cm := &schema.ConfigurationModel{
			Name: name,
			OriginalName: &name,
			UploadedBy: uploadedBy,
			MofFile: mofBytes,
			PreviousID: previousID,
		}
		err = db.CreateConfigurationModel(dbConn, cm)

		// 3. En cas d'échec de l'étape de création, on restore le nom de la configuration précédente (de versionnée à non versionnée) et on retourne l'opération en échec (et on log le détail du pourquoi en ERROR)
		if err != nil {
			if previousID != nil {
				_, _ = dbConn.Exec("UPDATE configuration_model SET name = ? WHERE id = ?", name, *previousID)
			}
			logs.WriteLogFile("[ERROR] Erreur CREATE configuration_model: " + err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Creation error"))
			return
		}

		// Met à jour le statut des agents liés à cette configuration
		var res sql.Result
		res, err = dbConn.Exec("UPDATE agents SET state = 'pending_apply' WHERE agent_id IN (SELECT agent_id FROM agent_configurations WHERE configuration_name = ?)", name)
		if err != nil {
			logs.WriteLogFile("[ERROR] Erreur UPDATE agents pending_apply: " + err.Error())
		} else {
			n, _ := res.RowsAffected()
			msg := fmt.Sprintf("[INFO] New configuration uploaded, impacted nodes: %d", n)
			logs.WriteLogFile(msg)
		}
	w.WriteHeader(http.StatusCreated)
}

// GET /api/v1/configuration_models/{id}
func GetConfigurationModelHandler(w http.ResponseWriter, r *http.Request) {
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbConn, err := db.OpenDB(dbCfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	cm, err := db.GetConfigurationModel(dbConn, id)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cm)
}

// GET /api/v1/configuration_models
func ListConfigurationModelsHandler(w http.ResponseWriter, r *http.Request) {
		dbCfg, err := db.LoadDBConfig("config.json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		dbConn, err := db.OpenDB(dbCfg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer dbConn.Close()

		// Support de l'option ?count=1
		if r.URL.Query().Get("count") == "1" {
			row := dbConn.QueryRow("SELECT COUNT(*) FROM configuration_model")
			var count int
			if err := row.Scan(&count); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]int{"count": count})
			return
		}

		list, err := db.ListConfigurationModels(dbConn)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if list == nil {
			list = make([]schema.ConfigurationModel, 0)
		}
		json.NewEncoder(w).Encode(list)
}

// PUT /api/v1/configuration_models/{id}
func UpdateConfigurationModelHandler(w http.ResponseWriter, r *http.Request) {
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbConn, err := db.OpenDB(dbCfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !auth.IsAdmin(r, dbConn) {
		http.Error(w, "Forbidden: admin only", http.StatusForbidden)
		return
	}
	name := r.FormValue("name")
	uploadedBy := r.FormValue("uploaded_by")
	file, _, err := r.FormFile("mof_file")
	var mofBytes []byte
	if err == nil {
		defer file.Close()
		mofBytes, _ = ioutil.ReadAll(file)
	}
	cm := &schema.ConfigurationModel{
		ID: id,
		Name: name,
		UploadedBy: uploadedBy,
		MofFile: mofBytes,
	}
	err = db.UpdateConfigurationModel(dbConn, cm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// DELETE /api/v1/configuration_models/{id}
func DeleteConfigurationModelHandler(w http.ResponseWriter, r *http.Request) {
	dbCfg, err := db.LoadDBConfig("config.json")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	dbConn, err := db.OpenDB(dbCfg)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer dbConn.Close()
	idStr := r.URL.Query().Get("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Récupère le nom de la configuration
	var configName string
	err = dbConn.QueryRow("SELECT name FROM configuration_model WHERE id = ?", id).Scan(&configName)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erreur lors de la récupération du nom de la configuration."))
		return
	}
	// Vérifie si un noeud utilise cette configuration via le nom
	var count int
	err = dbConn.QueryRow("SELECT COUNT(*) FROM agent_configurations WHERE configuration_name = ?", configName).Scan(&count)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Erreur lors de la vérification des associations de noeuds."))
		return
	}
	if count > 0 {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("This confguration is in use and cannot be deleted."))
		return
	}
	err = db.DeleteConfigurationModel(dbConn, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
