package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"crypto/sha256"
	"io"
	"strings"
	"fmt"
	"go-dsc-pull/internal/db"
	"go-dsc-pull/internal/schema"
)

// GetDscActionNodeHandlerWithId gère POST /PSDSCPullServer.svc/{id}/GetDscAction avec agentId déjà extrait
func GetDscActionNodeHandlerWithId(w http.ResponseWriter, r *http.Request, agentId string) {
		// Met à jour last_communication dans la table agents
		dbCfgUpdate, errUpdate := db.LoadDBConfig("config.json")
		if errUpdate == nil {
			database, err := db.OpenDB(dbCfgUpdate)
			if err == nil {
				_, err := database.Exec("UPDATE agents SET last_communication = CURRENT_TIMESTAMP WHERE agent_id = ?", agentId)
				if err != nil {
					log.Printf("[GETDSCACTION-NODE] Erreur update last_communication: %v", err)
				}
				database.Close()
			}
		}
	// Log du body et des headers reçus pour debug
	body, _ := io.ReadAll(r.Body)
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
						log.Printf("[GETDSCACTION-NODE] Nom de configuration trouvé pour agent %s : %s", agentId, name)
						configNames = append(configNames, name)
					}
				}
			}
			database.Close()
		}
	}

	// Vérifie qu'il existe au moins une configuration en base pour cet agent
	dbCfgCheck, errCheck := db.LoadDBConfig("config.json")
	if errCheck != nil {
		http.Error(w, "DB config error", http.StatusInternalServerError)
		return
	}
	dbConnCheck, errCheck := db.OpenDB(dbCfgCheck)
	if errCheck != nil {
		http.Error(w, "DB open error", http.StatusInternalServerError)
		return
	}
	defer dbConnCheck.Close()
	row := dbConnCheck.QueryRow("SELECT COUNT(*) FROM agent_configurations WHERE agent_id = ?", agentId)
	var count int
	if err := row.Scan(&count); err != nil {
		http.Error(w, "Configuration not found", http.StatusNotFound)
		return
	}
	// S'il n'y a aucune config, on ajoute l'ID de l'agent comme nom de config par défaut
	if count == 0 {
		configNames = append(configNames, agentId)
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
				// Gestion robuste du nom de configuration
				if configName == "" {
					if len(configNames) > 0 {
						configName = configNames[0]
					} else if agentId != "" {
						configName = agentId
					} else {
						http.Error(w, "Aucune configuration trouvée pour cet agent", http.StatusNotFound)
						return
					}
				}
				// Calcul du hash sur le bon fichier (lié à configName)
				   if cs.ChecksumAlgorithm == "SHA-256" && cs.Checksum != "" {
					   var hash string
					   // Récupère le MOF depuis la base
					   log.Printf("[GETDSCACTION-NODE] Recherche MOF en base pour configName='%s'", configName)
					   row := dbConnCheck.QueryRow("SELECT mof_file FROM configuration_model WHERE name = ? COLLATE NOCASE", configName)
					   var mofBytes []byte
					   if err := row.Scan(&mofBytes); err == nil {
						   hash = sha256SumHex(mofBytes)
					   } else {
						   allOk = false
					   }
					   log.Printf("[GETDSCACTION-NODE] Config=%s, Hash envoyé=%s, Hash calculé=%s", configName, cs.Checksum, hash)
						_, err := dbConnCheck.Exec("UPDATE configuration_model SET last_usage = CURRENT_TIMESTAMP WHERE name = ? COLLATE NOCASE", configName)
					    log.Printf("[GETDSCACTION-NODE] Mise à jour last_usage pour configName='%s', err=%v", configName, err)
						if hash != "" && strings.EqualFold(hash, cs.Checksum) {
						   status = "OK"
						   // Met à jour last_usage à CURRENT_TIMESTAMP pour cette configuration
						   if err != nil {
							   log.Printf("[GETDSCACTION-NODE] Erreur lors de la mise à jour de last_usage: %v", err)
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
			if len(configNames) > 0 {
				for _, name := range configNames {
					details = append(details, schema.DscActionDetail{
						ConfigurationName: name,
						Status:            "GetConfiguration",
					})
				}
			} else if agentId != "" {
				details = append(details, schema.DscActionDetail{
					ConfigurationName: agentId,
					Status:            "GetConfiguration",
				})
			} else {
				http.Error(w, "Aucune configuration trouvée pour cet agent", http.StatusNotFound)
				return
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
