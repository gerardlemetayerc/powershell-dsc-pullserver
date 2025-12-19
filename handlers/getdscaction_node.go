package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"crypto/sha256"
	"io"
	"strings"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
)

// GetDscActionNodeHandlerWithId gère POST /PSDSCPullServer.svc/{id}/GetDscAction avec agentId déjà extrait
func GetDscActionNodeHandlerWithId(w http.ResponseWriter, r *http.Request, agentId string) {
	// Log du body et des headers reçus pour debug
	body, _ := io.ReadAll(r.Body)
	log.Printf("[GETDSCACTION-NODE] AgentId=%s Body=%s", agentId, string(body))
	for k, v := range r.Header {
		log.Printf("[GETDSCACTION-NODE] Header: %s: %v", k, v)
	}
	r.Body = io.NopCloser(strings.NewReader(string(body)))

	// DEBUG : afficher le body dans la réponse HTTP (en plus du log)
	w.Header().Set("X-Debug-Client-Body", string(body))

	// Charger les noms de configuration depuis la base
	var configNames []string
	dbCfg, err := db.LoadDBConfig("config.json")
	if err == nil {
		database, err := db.OpenDB(dbCfg)
		if err == nil {
			rows, err := database.Query(`SELECT configuration_name FROM agent_configurations WHERE agent_id = ?`, agentId)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var name string
					if err := rows.Scan(&name); err == nil {
						configNames = append(configNames, name)
					}
				}
			}
			database.Close()
		}
	}

	mofPath := filepath.Join("configs", agentId+".mof")
	if _, err := os.Stat(mofPath); os.IsNotExist(err) {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}

	// Utiliser les schémas importés
	var cBody schema.ClientBody
	nodeStatus := "GetConfiguration"
	details := []schema.DscActionDetail{}

	if err := json.Unmarshal(body, &cBody); err == nil && len(cBody.ClientStatus) > 0 {
		// Si plusieurs objets dans ClientStatus, traiter chaque configuration
		allOk := true
		for _, cs := range cBody.ClientStatus {
			status := "GetConfiguration"
			configName := cs.ConfigurationName
			if configName == "" {
				configName = agentId
			}
			if cs.ChecksumAlgorithm == "SHA-256" && cs.Checksum != "" {
				mofPath := filepath.Join("configs", configName+".mof")
				if mofBytes, err := os.ReadFile(mofPath); err == nil {
					hash := sha256SumHex(mofBytes)
					if strings.EqualFold(hash, cs.Checksum) {
						status = "OK"
					} else {
						allOk = false
					}
				} else {
					allOk = false
				}
			} else {
				allOk = false
			}
			details = append(details, schema.DscActionDetail{
				ConfigurationName: configName,
				Status:            status,
			})
		}
		if allOk {
			nodeStatus = "OK"
		}
	} else {
		// fallback : comportement historique, une seule config
		if len(configNames) == 0 {
			configNames = append(configNames, agentId)
		}
		for _, name := range configNames {
			details = append(details, schema.DscActionDetail{
				ConfigurationName: name,
				Status:            "GetConfiguration",
			})
		}
	}

	resp := schema.DscActionResponse{
		NodeStatus: nodeStatus,
		Details:    details,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("ProtocolVersion", "2.0")
	json.NewEncoder(w).Encode(resp)
}

// sha256SumHex calcule le hash SHA-256 d'un tableau d'octets et retourne l'hexadécimal en majuscules
func sha256SumHex(data []byte) string {
	h := sha256.Sum256(data)
	return strings.ToUpper(fmt.Sprintf("%X", h[:]))
}
