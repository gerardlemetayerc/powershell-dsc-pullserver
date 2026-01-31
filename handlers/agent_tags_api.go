package handlers

import (
	"encoding/json"
	"net/http"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/logs"
)

// GET /api/v1/agents/{id}/tags : liste tous les tags clé/valeur d'un agent
func AgentTagsListHandler(w http.ResponseWriter, r *http.Request) {
	agentId := r.PathValue("id")
	       dbCfg, err := db.LoadDBConfig("config.json")
	       if err != nil {
			   logs.WriteLogFile("AgentTagsListHandler: DB config error: " + err.Error())
		       http.Error(w, "DB config error", http.StatusInternalServerError)
		       return
	       }
	       database, err := db.OpenDB(dbCfg)
	       if err != nil {
			   logs.WriteLogFile("AgentTagsListHandler: DB open error: " + err.Error())
		       http.Error(w, "DB open error", http.StatusInternalServerError)
		       return
	       }
	       defer database.Close()
	       tags, err := db.GetAgentTags(database, agentId)
	       if err != nil {
			   logs.WriteLogFile("AgentTagsListHandler: DB query error: " + err.Error())
		       http.Error(w, "DB query error", http.StatusInternalServerError)
		       return
	       }
	       w.Header().Set("Content-Type", "application/json")
	       _ = json.NewEncoder(w).Encode(tags)
}

// PUT /api/v1/agents/{id}/tags : ajoute une valeur à un tag clé pour un agent
func AgentTagsSetHandler(w http.ResponseWriter, r *http.Request) {
       agentId := r.PathValue("id")
       var req struct {
	       Key   string `json:"key"`
	       Value string `json:"value"`
       }
       if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Key == "" || req.Value == "" {
	       logs.WriteLogFile("AgentTagsSetHandler: Invalid key/value: " + err.Error())
	       http.Error(w, "Invalid key/value", http.StatusBadRequest)
	       return
       }
       dbCfg, err := db.LoadDBConfig("config.json")
       if err != nil {
	       logs.WriteLogFile("AgentTagsSetHandler: DB config error: " + err.Error())
	       http.Error(w, "DB config error", http.StatusInternalServerError)
	       return
       }
       database, err := db.OpenDB(dbCfg)
       if err != nil {
	       logs.WriteLogFile("AgentTagsSetHandler: DB open error: " + err.Error())
	       http.Error(w, "DB open error", http.StatusInternalServerError)
	       return
       }
       defer database.Close()
       if err := db.SetAgentTag(database, agentId, req.Key, req.Value); err != nil {
	       logs.WriteLogFile("AgentTagsSetHandler: DB update error: " + err.Error())
	       http.Error(w, "DB update error", http.StatusInternalServerError)
	       return
       }
       w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/agents/{id}/tags : supprime une valeur précise d'un tag clé pour un agent
func AgentTagsDeleteHandler(w http.ResponseWriter, r *http.Request) {
       agentId := r.PathValue("id")
       key := r.URL.Query().Get("key")
       value := r.URL.Query().Get("value")
       if key == "" || value == "" {
	       logs.WriteLogFile("AgentTagsDeleteHandler: Missing key or value")
	       http.Error(w, "Missing key or value", http.StatusBadRequest)
	       return
       }
       dbCfg, err := db.LoadDBConfig("config.json")
       if err != nil {
	       logs.WriteLogFile("AgentTagsDeleteHandler: DB config error: " + err.Error())
	       http.Error(w, "DB config error", http.StatusInternalServerError)
	       return
       }
       database, err := db.OpenDB(dbCfg)
       if err != nil {
	       logs.WriteLogFile("AgentTagsDeleteHandler: DB open error: " + err.Error())
	       http.Error(w, "DB open error", http.StatusInternalServerError)
	       return
       }
       defer database.Close()
       if err := db.DeleteAgentTag(database, agentId, key, value); err != nil {
	       logs.WriteLogFile("AgentTagsDeleteHandler: DB delete error: " + err.Error())
	       http.Error(w, "DB delete error", http.StatusInternalServerError)
	       return
       }
       w.WriteHeader(http.StatusNoContent)
}
