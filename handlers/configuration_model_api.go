package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"go-dsc-pull/internal/db"
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
	property := r.FormValue("property")
	value := r.FormValue("value")
	file, _, err := r.FormFile("mof_file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Missing MOF file"))
		return
	}
	defer file.Close()
	mofBytes, err := ioutil.ReadAll(file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	cm := &db.ConfigurationModel{
		Property: property,
		Value: value,
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
	list, err := db.ListConfigurationModels(dbConn)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if list == nil {
		list = make([]db.ConfigurationModel, 0)
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
	property := r.FormValue("property")
	value := r.FormValue("value")
	file, _, err := r.FormFile("mof_file")
	var mofBytes []byte
	if err == nil {
		defer file.Close()
		mofBytes, _ = ioutil.ReadAll(file)
	}
	cm := &db.ConfigurationModel{
		ID: id,
		Property: property,
		Value: value,
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
