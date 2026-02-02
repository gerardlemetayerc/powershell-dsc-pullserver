package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"fmt"
	"go-dsc-pull/internal/utils"
	"go-dsc-pull/internal/db"
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
	filename := header.Filename
	if len(filename) < 5 || filename[len(filename)-4:] != ".mof" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Le fichier doit être au format .mof"))
		return
	}
	       // Calcule le nom du modèle (nom du fichier sans .mof)
	       name := strings.ToLower(filename[:len(filename)-4])
	       mofBytes, err := ioutil.ReadAll(file)
	       if err != nil {
		       w.WriteHeader(http.StatusInternalServerError)
		       return
	       }
		       // Vérifie si une config existe déjà avec ce nom
		       var existingId int64
		       var existingDate string
		       err = dbConn.QueryRow("SELECT id, upload_date FROM configuration_model WHERE name = ? ORDER BY upload_date LIMIT 1", name).Scan(&existingId, &existingDate)
		       if err == nil && existingId > 0 {
			       // Renomme l'ancienne config : <nom>-<dateuploadinitial>
			       safeDate := existingDate
			       safeDate = strings.ReplaceAll(safeDate, ":", "")
			       safeDate = strings.ReplaceAll(safeDate, "-", "")
			       safeDate = strings.ReplaceAll(safeDate, " ", "_")
			       newName := name + "-" + safeDate
			       _, _ = dbConn.Exec("UPDATE configuration_model SET name = ? WHERE id = ?", newName, existingId)
		       }

			   // Met à jour le statut des agents liés à cette configuration
			   res, err := dbConn.Exec("UPDATE agents SET state = 'pending_apply' WHERE agent_id IN (SELECT agent_id FROM agent_configurations WHERE LOWER(configuration_name) = ?)", name)
			   if err != nil {
				   // log erreur
				   logs.WriteLogFile("[ERROR] Erreur UPDATE agents pending_apply: " + err.Error())
			   } else {
				   n, _ := res.RowsAffected()
				   msg := fmt.Sprintf("[INFO] New configuration uploaded, impacted nodes: %d", n)
				   logs.WriteLogFile(msg)
			   }
	// Récupère le username depuis le JWT (claim "sub")
	uploadedBy := "?"
	if r.Context().Value("user") != nil {
		if sub, ok := r.Context().Value("user").(string); ok {
			uploadedBy = sub
		}
	} else if auth := r.Header.Get("Authorization"); len(auth) > 7 {
		// Décodage manuel du JWT si le middleware ne pose pas le contexte
		tokenStr := auth[7:]
		// Utilise la même clé que le middleware
		secret := []byte("supersecretkey")
		token, _ := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) { return secret, nil })
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			if sub, ok := claims["sub"].(string); ok {
				uploadedBy = sub
			}
		}
	}
	cm := &schema.ConfigurationModel{
		Name: name,
		UploadedBy: uploadedBy,
		MofFile: mofBytes,
	}
	err = db.CreateConfigurationModel(dbConn, cm)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
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
	if !utils.IsAdmin(r, dbConn) {
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
	err = db.DeleteConfigurationModel(dbConn, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
